import os

CURRENT_VERSION_ID = os.environ.get('CURRENT_VERSION_ID', None)
if os.environ.get('SERVER_SOFTWARE', '').startswith('Google App Engine'):
	DEVELOPMENT = False
else:
	DEVELOMENT = True

PRODUCTION = not DEVELOPMENT
DEBUG = DEVELOPMENT

DEFAULT_DB_LIMIT = 64

#####################################################################
# Cient modules, use by the build.py script.
#####################################################################
STYLES = [
	'/static/less/style.less',
}

