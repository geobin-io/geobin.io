'use strict';

/* https://github.com/angular/protractor/blob/master/docs/getting-started.md */

describe('my app', function() {

  browser.get('index.html');

  it('should automatically redirect to / when location hash/fragment is empty', function() {
    expect(browser.getLocationAbsUrl()).toMatch("/");
  });


  describe('bin', function() {

    beforeEach(function() {
      browser.get('index.html#/asdf');
    });

    it('should render bin when user navigates to /asdf', function() {
      expect(element.all(by.css('[ng-view] p')).first().getText()).
        toMatch(/partial for bin/);
    });

  });
});
