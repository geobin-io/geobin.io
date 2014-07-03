# Geobin Client

[![Development Dependencies](https://david-dm.org/esripdx/geobin.io/dev-status.svg)](https://david-dm.org/esripdx/geobin.io#info=devDependencies)

Browser client for Geobin.

### Install Dependencies

We have two kinds of dependencies in this project: tools and angular framework code. The tools help us manage and test the application.

* We get the tools we depend upon via `npm`, the [node package manager][npm].
* We get the angular code via `bower`, a [client-side code package manager][bower].

We have preconfigured `npm` to automatically run `bower` so we can simply do:

```bash
npm install
```

Behind the scenes this will also call `./node_modules/bower/bin/bower install`. You should find that you have two new folders in your project.

* `/node_modules` - contains the npm packages for the tools we need
* `/static/app/components` - contains the angular framework files

### Run the Application

Make sure you've also installed all server dependencies (`go get`) before starting the application.

```bash
make run
```

Now browse to the app at `http://localhost:8080/`.

## Directory Layout

    static/app/         --> all of the files to be used in production
      css/              --> css files
        app.css         --> default stylesheet
      img/              --> image files
      js/               --> javascript files
        app.js          --> application
        controllers.js  --> application controllers
        directives.js   --> application directives
        filters.js      --> custom angular filters
        services.js     --> custom angular services
      partials/             --> angular view partials (partial html templates)
        home.html
        bin.html

    test/               --> test config and source files
      protractor-conf.js    --> config file for running e2e tests with Protractor
      e2e/                  --> end-to-end specs
        scenarios.js
      karma.conf.js         --> config file for running unit tests with Karma
      unit/                 --> unit level specs/tests
        controllersSpec.js      --> specs for controllers
        directivessSpec.js      --> specs for directives
        filtersSpec.js          --> specs for filters
        servicesSpec.js         --> specs for services

## Testing

### Running Unit Tests

Unit tests are written in [Jasmine][jasmine], which we run with the [Karma Test Runner][karma].

* the configuration is found at `test/karma.conf.js`
* the unit tests are found in `test/unit/`.

The easiest way to run the unit tests is to use the supplied npm script:

```bash
npm test
```

This script will start the Karma test runner to execute the unit tests. This will do a single run of the tests using PhantomJS and then exit.

Karma can also sit and watch the source and test files for changes and then re-run the tests whenever any of them change using Chrome, Firefox, and PhantomJS browers. This is the recommended strategy when developing locally; if your unit tests are being run every time you save a file then you receive instant feedback on any changes that break the expected code functionality.

To run tests continuously with Chrome, Firefox, and PhantomJS, you can use the following supplied npm script:

```bash
npm run test-forever
```

### End to end testing

End-to-end tests are written in [Jasmine][jasmine] and run with the [Protractor][protractor] End-to-End test runner.  It uses native events and has special features for Angular applications.

* the configuration is found at `test/protractor-conf.js`
* the end-to-end tests are found in `test/e2e/`

Protractor simulates interaction with our web app and verifies that the application responds correctly. Therefore, our web server needs to be serving up the application, so that Protractor can interact with it.

```bash
make run
```

In addition, since Protractor is built upon WebDriver we need to install it by running this script:

```bash
npm run update-webdriver
```

This will download and install the latest version of the stand-alone WebDriver tool.

Once you have ensured that the development web server hosting our application is up and running and WebDriver is updated, you can run the end-to-end tests using the supplied npm script:

```bash
npm run test-e2e
```

This script will execute the end-to-end tests against the application being hosted on the development server.

## Updating Dependencies

You can update the tool dependencies by running:

```bash
npm update
```

This will find the latest versions that match the version ranges specified in the `package.json` file.

You can update the Angular dependencies by running:

```bash
bower update
```

This will find the latest versions that match the version ranges specified in the `bower.json` file.

[bower]: http://bower.io
[npm]: https://www.npmjs.org/
[node]: http://nodejs.org
[protractor]: https://github.com/angular/protractor
[jasmine]: http://pivotal.github.com/jasmine/
[karma]: http://karma-runner.github.io
