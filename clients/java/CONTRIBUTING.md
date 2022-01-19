# CONTRIBUTING

### Publishing to Maven Central

- Increment the version number in `build.gradle` and `Config.java`
- Set all the `gradle.properties` variables referenced in `build.gradle`. Follow this guide to get help: https://madhead.me/posts/no-bullshit-maven-publish/
- Run `./gradlew publish`
- Update the config of recommended and deprecated versions in `services/data/clientversion/clientversion.go`
