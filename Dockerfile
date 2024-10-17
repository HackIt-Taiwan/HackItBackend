FROM golang:1.22 as build
WORKDIR /
COPY . .
RUN go build -o /server .

FROM scratch
COPY --from=build /server /server
EXPOSE 5000
CMD ["/server"]