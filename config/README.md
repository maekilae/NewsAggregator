### Application configuration

```json
{
  "log": {
    "log_file_level": "log level for file output e.g. debug",
    "console_level": "log level for console output e.g. info"
  },
  "api": {
    "enabled": true,
    "bearer_token": true, whether or not to enable bearer token authentication
    "path": "/api/v1",
    "port": 8080,
    "rate_limit": integer value for rate limit, 0 to disable
  },
  "classifier": {
    "enabled": true,
    "model_name": "model name used with provider e.g. mistral_small_latest",
    "provider_endpoint": "https://api.mistral.ai/",
    "token_env": "environment variable for api token e.g. MISTRAL_API_TOKEN"
  }
}
```

### Feed provider configuration

```json
{
  "name": "Used for output e.g. SVT News",
  "url": "url to feed e.g. https://www.svt.se/rss.xml"
}
```

### Classifier configuration

If you want the classifier to use a different system message you can provide it by creating an SYSMSG.md file in the config folder i.e. `config/SYSMSG.md`

Default system message can be found in `config/DEFAULT_SYSMSG.md`, note that the file is not read.
