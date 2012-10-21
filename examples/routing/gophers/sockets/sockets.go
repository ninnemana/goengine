// Copyright 2010 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// Package websocket implements the WebSocket protocol for the Twister server.
// 
// WARNING: This packages has not been updated to the latest WebSocket
// specification.
package sockets

/*
import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	//"github.com/garyburd/twister/web"
	"io"
	"net"
	"net/http"
	"strings"
)

const (
	headerSecWebSocketKey1     = "Sec-Websocket-Key1"
	headerSecWebSocketKey2     = "Sec-Websocket-Key2"
	headerSecWebSocketProtocol = "Sec-Websocket-Protocol"
)

type Conn struct {
	conn    net.Conn
	br      *bufio.Reader
	bw      *bufio.Writer
	hasMore bool
}

func (conn *Conn) Close() error {
	return conn.conn.Close()
}

// ReadMessage reads a message from the client. The message is returned in one
// or more chunks. hasMore is set to false on the last chunk of the message.
// If the message fits in the read buffer size specified in the call to
// Upgrade, then the message is guaranteed to be returned in a single chunk.
// The returned chunk points to the internal state of the connection and is only
// valid until the next call to ReadMessage.
func (conn *Conn) ReadMessage() (chunk []byte, hasMore bool, err error) {
	// Support text framing for now. Revisit after browsers support framing
	// described in later specs.

	if !conn.hasMore {
		c, err := conn.br.ReadByte()
		if err != nil {
			return nil, false, err
		}
		if c != 0 {
			return nil, false, errors.New("twister.websocket: unexpected framing.")
		}
	}

	p, err := conn.br.ReadSlice(0xff)
	switch err {
	case bufio.ErrBufferFull:
		conn.hasMore = true
	case nil:
		p = p[:len(p)-1]
		conn.hasMore = false
	default:
		return nil, false, err
	}
	return p, conn.hasMore, nil
}

// WriteMessage write a message to the client. The message cannot contain the
// bytes with value 0 or 255.
func (conn *Conn) WriteMessage(p []byte) error {
	// Support text framing for now. Revisit after browsers support framing
	// described in later specs.
	conn.bw.WriteByte(0)
	conn.bw.Write(p)
	conn.bw.WriteByte(0xff)
	return conn.bw.Flush()
}

// webSocketKey returns the key bytes from the specified websocket key header.
func webSocketKey(req *http.Request, name string) (key []byte, err error) {
	s := req.Header.Get(name)
	if s == "" {
		return key, errors.New("twister.websocket: missing key")
	}
	var n uint32 // number formed from decimal digits in key
	var d uint32 // number of spaces in key
	for i := 0; i < len(s); i++ {
		b := s[i]
		if b == ' ' {
			d += 1
		} else if '0' <= b && b <= '9' {
			n = n*10 + uint32(b) - '0'
		}
	}
	if d == 0 || n%d != 0 {
		return nil, errors.New("twister.websocket: bad key")
	}
	key = make([]byte, 4)
	binary.BigEndian.PutUint32(key, n/d)
	return key, nil
}

// Upgrade upgrades the HTTP connection to the WebSocket protocol. The 
// caller is responsible for closing the returned connection.
func Upgrade(req *http.Request, readBufSize, writeBufSize int, header web.Header) (conn *Conn, err error) {

	if req.Method != "GET" {
		http.StatusInternalServerError
		//req.Respond(web.StatusMethodNotAllowed)
		return nil, errors.New("twister.websocket: bad request method")
	}

	origin := req.Header.Get(web.HeaderOrigin)
	if origin == "" {
		req.Respond(web.StatusBadRequest)
		return nil, errors.New("twister.websocket: origin missing")
	}

	connection := strings.ToLower(req.Header.Get(web.HeaderConnection))
	if connection != "upgrade" {
		req.Respond(web.StatusBadRequest)
		return nil, errors.New("twister.websocket: connection header missing or wrong value")
	}

	upgrade := strings.ToLower(req.Header.Get(web.HeaderUpgrade))
	if upgrade != "websocket" {
		req.Respond(web.StatusBadRequest)
		return nil, errors.New("twister.websocket: upgrade header missing or wrong value")
	}

	key1, err := webSocketKey(req, headerSecWebSocketKey1)
	if err != nil {
		req.Respond(web.StatusBadRequest)
		return nil, err
	}

	key2, err := webSocketKey(req, headerSecWebSocketKey2)
	if err != nil {
		req.Respond(web.StatusBadRequest)
		return nil, err
	}

	netConn, br, err := req.Responder.Hijack()
	if err != nil {
		return nil, err
	}

	defer func() {
		if netConn != nil {
			netConn.Close()
		}
	}()

	var r io.Reader
	if br.Buffered() > 0 {
		buf, _ := br.Peek(br.Buffered())
		r = io.MultiReader(bytes.NewBuffer(buf), netConn)
	} else {
		r = netConn
	}

	br = bufio.NewReaderSize(r, readBufSize)
	bw := bufio.NewWriterSize(netConn, writeBufSize)

	key3 := make([]byte, 8)
	if _, err := io.ReadFull(br, key3); err != nil {
		req.Respond(web.StatusBadRequest)
		return nil, err
	}

	hash := md5.New()
	hash.Write(key1)
	hash.Write(key2)
	hash.Write(key3)
	response := hash.Sum(nil)

	// TODO: handle tls
	location := "ws://" + req.URL.Host + req.URL.RequestURI()
	protocol := req.Header.Get(headerSecWebSocketProtocol)

	h := make(web.Header)
	for k, v := range header {
		h[k] = v
	}
	h.Set("Upgrade", "WebSocket")
	h.Set("Connection", "Upgrade")
	h.Set("Sec-Websocket-Location", location)
	h.Set("Sec-Websocket-Origin", origin)
	if len(protocol) > 0 {
		h.Set("Sec-Websocket-Protocol", protocol)
	}

	if _, err := bw.WriteString("HTTP/1.1 101 WebSocket Protocol Handshake\r\n"); err != nil {
		return nil, err
	}

	if err := h.WriteHttpHeader(bw); err != nil {
		return nil, err
	}

	if _, err := bw.Write(response); err != nil {
		return nil, err
	}

	if err := bw.Flush(); err != nil {
		return nil, err
	}

	conn = &Conn{netConn, br, bw, false}
	netConn = nil
	return conn, nil
}
*/
