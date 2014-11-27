module.exports = function(config){
  config.set({
    basePath: '../',
    files: [
      'static/components/angular/angular.js',
      'static/components/angular-ui-router/release/angular-ui-router.js',
      'static/components/angular-mocks/angular-mocks.js',
      'static/components/tinystore/tinystore.min.js',
      'static/app/index.js',
      'static/app/**/*.js',
      'test/unit/**/*.js'
    ],
    autoWatch: true,
    frameworks: ['jasmine'],
    browsers: ['Chrome','Firefox','PhantomJS'],
    plugins: [
      'karma-chrome-launcher',
      'karma-firefox-launcher',
      'karma-phantomjs-launcher',
      'karma-jasmine',
      'karma-coverage',
      'karma-mocha-reporter'
    ],
    junitReporter: {
      outputFile: 'test_out/unit.xml',
      suite: 'unit'
    },
    reporters: ['mocha', 'coverage'],
    preprocessors: {
      // source files, that you wanna generate coverage for
      // do not include tests or libraries
      // (these files will be instrumented by Istanbul)
      'static/app/*.js': ['coverage']
    },
    coverageReporter: {
      reporters:[
        {type: 'html', dir:'coverage/'},
        {type: 'text'}
      ]
    }
  });
};
