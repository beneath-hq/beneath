# CONTRIBUTING

### Publishing to Maven Central

1. Increment the version number in `build.gradle`
2. Set all the `gradle.properties` variables referenced in `build.gradle`. Follow this guide to get help: https://madhead.me/posts/no-bullshit-maven-publish/
3. Run `./gradlew publish`
