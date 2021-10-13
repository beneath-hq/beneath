Note on Maven plugins in `pom.xml`:
- Plugins inside `<pluginManagement></pluginManagement>` tags don't get run on `mvn install`
- Plugins are flaky when you try to execute them independently, outside of the default lifecycle. E.g. running `build-helper:add-source` returns an error. But running `mvn install`, which runs through the default lifecyle, executes `build-helper:add-source` successfully.
