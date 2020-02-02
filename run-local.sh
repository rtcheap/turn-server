cd cmd
go build
cd ..
mv cmd/cmd turn-server

export SERVICE_NAME='localhost'
export SERVICE_PORT='8081'
export SERVICE_REGISTRY_URL='http://localhost:8080'

export JWT_ISSUER='rtcheap'
export JWT_SECRET='password'

export JAEGER_SERVICE_NAME='turn-server'
export JAEGER_SAMPLER_TYPE='const'
export JAEGER_SAMPLER_PARAM=1
export JAEGER_REPORTER_LOG_SPANS='1'

./turn-server