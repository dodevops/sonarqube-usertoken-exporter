# Sonarqube usertoken exporter

A prometheus exporter that exports currently issued user tokens of a Sonarqube instance
and when they expire.

## Usage

Start the container (e.g. using podman or docker) to launch the exporter locally.

    podman run --rm -p 8081:8081 -e EXPORTER_URL=https://sonarqube.company.com -e EXPORTER_TOKEN=verysecretadmintoken ghcr.io/dodevops/sonarqube-usertoken-exporter:latest

After that, the metrics can be retrieved using http://localhost:8081/metrics

The following environment variables are required to configure the exporter:

* EXPORTER_URL: URL to the sonarqube instance
* EXPORTER_TOKEN: A user token with admin privileges

And these are optional for further configuration:

* EXPORTER_LOGLEVEL: Level to use when logging out information [INFO]
* EXPORTER_PORT: Port to listen on [8081]
* EXPORTER_INTERVAL: Interval to sleep between metric gathering in minutes [60]

## Metrics

The following metrics are exporter:

### sonarqube_user_tokens_expiration_date_seconds

The expiration of a user token as a unix epoch.

The following labels are provided:

* user: The username that the token is bound to
* token: Name of the token
* type: User token type. Can be USER_TOKEN, GLOBAL_ANALYSIS_TOKEN or PROJECT_ANALYSIS_TOKEN
* project_key: Assigned project key if the key is of the PROJECT_ANALYSIS_TOKEN type
* is_expired: A boolean "true"/"false" whether the user token has expired

### sonarqube_user_tokens_creation_date_seconds

The creation date of a user token as a unix epoch

The following labels are provided:

* user: The username that the token is bound to
* token: Name of the token
* type: User token type. Can be USER_TOKEN, GLOBAL_ANALYSIS_TOKEN or PROJECT_ANALYSIS_TOKEN
* project_key: Assigned project key if the key is of the PROJECT_ANALYSIS_TOKEN type
* is_expired: A boolean "true"/"false" whether the user token has expired
