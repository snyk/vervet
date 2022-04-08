# S3 Local Testing

Prerequisites:

* LocalStack [CLI installed](https://docs.localstack.cloud/get-started/#installation)
  * `python3 -m pip install localstack`
* If you need to make changes
`localstack config validate` to verify docker-compose setup

To validate in Docker-compose `docker-compose up -d`
* `localhost:4566/health` will show status of all mimicked services