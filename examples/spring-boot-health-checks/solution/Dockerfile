# Stage 1: build the Spring Boot executable jar with Gradle + JDK 21.
FROM gradle:8.14-jdk21 AS builder
WORKDIR /home/gradle/src
COPY --chown=gradle:gradle . .
RUN gradle bootJar --no-daemon

# Stage 2: minimal JRE runtime, running as a non-root user (restricted Pod Security).
# The eclipse-temurin image already provides a non-root UID 1000.
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /home/gradle/src/build/libs/*.jar /app/app.jar
USER 1000
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "/app/app.jar"]
