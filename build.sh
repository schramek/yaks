docker build -t docker-dev.art.intern/bamf/cnc/cnc-yaks-base:20220722 .

DOCKER_CONTAINER_ID=$(docker run -it -d docker-dev.art.intern/bamf/cnc/cnc-yaks-base:20220722)

docker cp $DOCKER_CONTAINER_ID:/app/yaks ./yaks

docker stop $DOCKER_CONTAINER_ID

echo "Yaks Executable was generated."