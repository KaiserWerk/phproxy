# phproxy
phproxy is a fast & tiny app proxy which, at a foundational level, hides your PHP app and runs all requests through itself first instead.
That way you can safely run your unexposed PHP app locally while profiting from increased performance (asset & page content caching), greatly simplified TLS certificate setup and URL rewriting (even if the underlying server does not enable it), all configurable with a small YAML configuration file.

### Feature overview

1. Rewrite Urls in Links and redirect header calls
2. test run mode runs a quick test if all routes are working properly
3. cache static assets and content pages for increased performance
4. add a cert+key to enable tls or enable autocert for domain from letsencrypt
5. rewrite urls
    1. single
    2. from .htaccess file
    3. custom rule/custom html content

### Configuration

tbd

### Usage

tbd

### Logs

tbd


