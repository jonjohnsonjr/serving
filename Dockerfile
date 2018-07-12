# In knative
FROM gcr.io/jonjohnson-test/kythe/bins as bins

# Index the code, producing the serving_table
FROM golang as indexer
COPY --from=bins /out/gotool /gotool
COPY --from=bins /out/go_indexer /go_indexer
COPY --from=bins /out/write_tables /write_tables
COPY index.sh /index.sh
COPY . /go/src/github.com/knative/serving
WORKDIR /go/src/github.com/knative/serving
RUN bash /index.sh ./...

# We expect the serving table to be in /workspace/index
FROM gcr.io/jonjohnson-test/kythe/dist as dist

# Final image, serves the index via web ui
FROM ubuntu
COPY --from=dist /src/kythe/web/ui/resources/public /public
COPY --from=bins /out/http_server /http_server
RUN cp -r /workspace/index /index
EXPOSE 8080
CMD ["/http_server", "--listen", ":8080", "--public_resources", "/public", "--serving_table", "/index"]
