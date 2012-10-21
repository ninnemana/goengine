define(
	['jquery', '/js/libs/modernizr.min.js', '/js/plugins.js'],
	function($, Modernizr, bootstrap){
		console.log($);
		console.log(Modernizr);
		console.log(bootstrap);
		return $;
	}
);