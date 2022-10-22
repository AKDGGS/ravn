## Build
FROM golang:1.19-alpine AS build
WORKDIR /ravn
COPY . .
RUN go build -ldflags="-s -w"
RUN ./ravn genera data/taxon\ genera\ files/*.xlsx
RUN ./ravn references data/taxon\ references/*.txt
RUN ./ravn species -images data/taxon\ pictures \
	data/taxon\ species\ files/*.xlsx

## Deploy
FROM alpine:latest
WORKDIR /ravn
COPY --from=build /ravn/ravn .
COPY --from=build /ravn/species.bleve species.bleve
COPY --from=build /ravn/references.bleve references.bleve
COPY --from=build /ravn/genera.bleve genera.bleve
COPY ["data/taxon pictures", "images"]
EXPOSE 8080/tcp
CMD ./ravn start -listen :8080
