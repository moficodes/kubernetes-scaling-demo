# for loop from i = 1 to 10

for i in $(seq -w 1 40)
    do
        docker run -p 80$i:8080 \
        -d --name instance$i \
        -v "$HOME/.config/gcloud:/gcp/config:ro" \
        -v /gcp/config/logs \
        --env CLOUDSDK_CONFIG=/gcp/config \
        --env GOOGLE_APPLICATION_CREDENTIALS=/gcp/config/application_default_credentials.json \
        --env PROJECT_ID=mofilabs-next-demo-02 \
        -e HOSTNAME=instance$i \
        moficodes/instance:v1.5.0
    done
