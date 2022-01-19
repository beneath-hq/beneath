# Create builder image
FROM gradle:7.3.3-jdk17-alpine as builder
WORKDIR /app
COPY build.gradle settings.gradle .
COPY src src
RUN gradle --no-daemon build

# Create run image
FROM openjdk:17-alpine
COPY --from=builder /app/build/libs/cdc-generator-1.0.0-SNAPSHOT.jar app.jar
ENTRYPOINT java -jar app.jar
