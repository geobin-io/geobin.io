module.exports = function(config){
  config.set({
    basePath: '../',
    files: [
      'static/app/components/angular/angular.js',
      'static/app/components/angular-route/angular-route.js',
      'static/app/components/angular-mocks/angular-mocks.js',
      'static/app/components/tinystore/tinystore.min.js',
      'static/app/js/**/*.js',
      'test/unit/**/*.js'
    ],
    autoWatch: true,
    frameworks: ['jasmine'],
    browsers: ['Chrome'],
    plugins: [
      'karma-chrome-launcher',
      'karma-firefox-launcher',
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
      'static/app/js/*.js': ['coverage']
    },
    coverageReporter: {
      reporters:[
        {type: 'html', dir:'coverage/'},
        {type: 'text'}
      ]
    }
  });
};
