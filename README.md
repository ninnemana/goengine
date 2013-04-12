# GoEngine Boilerplate [![Build Status](https://travis-ci.org/ninnemana/goengine.png?branch=non-appengine)](https://travis-ci.org/ninnemana/goengine)

GoEngine Boilerplate is combination of serveral different repositories, some made for Go some not.

The project is built using with Sinatra style routing and MySQL support. HTML Structure and design using h5bp(http://html5boilerplate.com) and Twitter Bootstrap(http://twitter.github.com/bootstrap). Javascript templating is done with mustache.js, but has not been baked into requireJS. The javascript component incorporates RequireJS for dependency management.

**IMPORTANT**: If you **__are__** using Google AppEngine, use the [_gae_](https://github.com/ninnemana/goengine/tree/gae) branch.

Deploying
-----------

You can deploy this app manually or using Capistrano.  If you choose to use Capistrano, here's an example to prompt for populating the ConnectionString so you can keep it out of your public repo for testing purposes:

```ruby
default_run_options[:pty] = true
ssh_options[:forward_agent] = true


set :application, "ApplicationName"
set :appname,     "application-name"
set :repository,  "git@github.com:youraccount/yourrepo.git"

set :scm, :git
set :scm_passphrase, ""
set :user, "linuxuser"

role :web, "server1.address.com", "server2.address.com"
role :app, "server1.address.com", "server2.address.com"

set :deploy_to, "/home/#{user}/#{application}"
set :deploy_via, :remote_cache

set :use_sudo, false
set :sudo_prompt, ""
set :normalize_asset_timestamps, false

after :deploy, "deploy:goget", "db:configure", "deploy:compile", "deploy:stop", "deploy:restart"

namespace :db do
  desc "set database Connection String"
  task :configure do
    set(:database_username) { Capistrano::CLI.ui.ask("Database Username:") }
  
    set(:database_password) { Capistrano::CLI.password_prompt("Database Password:") }

    db_config = <<-EOF
      package database

      const (
        db_proto = "tcp"
        db_addr  = "databaseaddress:3306"
        db_user  = "#{database_username}"
        db_pass  = "#{database_password}"
        db_name  = "databasename"
      )
    EOF
    run "mkdir -p #{deploy_to}/current/helpers/database"
    put db_config, "#{deploy_to}/current/helpers/database/ConnectionString.go"
  end
end
namespace :deploy do
  task :goget do
    run "/usr/local/go/bin/go get github.com/ziutek/mymysql/native"
    run "/usr/local/go/bin/go get github.com/ziutek/mymysql/mysql"
  end
  task :compile do
    run "GOOS=linux GOARCH=amd64 CGO_ENABLED=0 /usr/local/go/bin/go build -o #{deploy_to}/current/#{appname} #{deploy_to}/current/index.go"
  end
  task :start do ; end
  task :stop do 
      kill_processes_matching "#{appname}"
  end
  task :restart do
    restart_cmd = "#{current_release}/#{appname} -http=127.0.0.1:8080 -path=#{deploy_to}/current/"
    run "nohup sh -c '#{restart_cmd} &' > nohup.out"
  end
end

def kill_processes_matching(name)
  run "ps -ef | grep #{name} | grep -v grep | awk '{print $2}' | sudo xargs kill -2 || echo 'no process with name #{name} found'"
end

```

Depending on the server you're on, you may need to define the path of the application as a flag in the deployment script or when running manually on the server. The reason for that is the server may have issues pathing the static files in the project. The example in the script above should help get the deployment working properly with the right static file paths.

Mustache.js
-----------

mustache.js has been converted to use [[ ]] as delimiters so it can play nice with golang's html/template package.

Issues
-----------

There is currently (1.7.0) an issue with passing routes with spaces on the App Engine dev_appserver.py. The issue does not seem to exist on the live server. We have found that making a small change to /google/appengine/ext/go/__init__.py will resolve this issue.

Remove from line 513:
```
request_uri = env['PATH_INFO']
```

Replace with:
```
request_uri = env['_AH_ENCODED_SCRIPT_NAME']
```

Contributors
-----------

**Alex Ninneman**

+ http://twitter.com/ninnemana
+ http://github.com/ninnemana

**Jessica Janiuk**

+ http://twitter.com/janiukjf
+ http://github.com/janiukjf
