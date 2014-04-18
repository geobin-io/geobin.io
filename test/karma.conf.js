module.exports = function(config){
  config.set({
    basePath: '../',
    files: [
      'static/app/components/angular/angular.js',
      'static/app/components/angular-route/angular-route.js',
      'static/app/components/angular-mocks/angular-mocks.js',
      'static/app/js/**/*.js',
      'test/unit/**/*.js'
    ],
    autoWatch: true,
    frameworks: ['jasmine'],
    browsers: ['Chrome'],
    plugins: [
      'karma-chrome-launcher',
      'karma-firefox-launcher',
      'karma-jasmine'
    ],
    junitReporter: {
      outputFile: 'test_out/unit.xml',
      suite: 'unit'
    }
  });
};