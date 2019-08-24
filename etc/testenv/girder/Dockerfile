FROM girder/girder:latest

RUN apt-get update && apt-get install --assume-yes netcat

ADD bootstrap.py /bootstrap.py
ADD entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]
