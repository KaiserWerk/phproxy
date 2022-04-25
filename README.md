# phproxy (Work in progress!)
phproxy is a fast & tiny app proxy which, at a foundational level, hides your PHP 
app (or static web app) and runs all requests through itself first instead.
That way you can safely run your unexposed PHP app locally while profiting from 
increased performance (asset & page content caching), greatly simplified TLS 
certificate setup and URL rewriting (even if the underlying server does not 
enable it), all configurable with a small YAML configuration file.

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

The YAML content of a configuration file might look like this:

```yaml
phproxy:
    bind_addr: :8881
    tls:
        key_file: /opt/certs/key.pem
        cert_file: /opt/certs/key.pem
        letsencrypt:
            enabled: false
            domain: mydomain.com
            email: me@mydomain.com
            accept_tos: true
apps:
    - name: testapp
      caching:
        level: 1
        static_assets:
            - assets/css/style.css
            - assets/js/scripts.js
            - assets/images
      url_rewriting:
        custom:
            /some/{id:[0-9]+}/{action}: some.php?action={action}&id={id}
        htaccess_files:
            - .htaccess
            - subfolder/.htaccess
```

### Usage

tbd

### Logs

tbd


