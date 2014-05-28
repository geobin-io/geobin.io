'use strict';

/* jasmine specs for services go here */

describe('basemaps', function() {
  beforeEach(module('Geobin.services'));

  var basemaps;
  beforeEach(function() {
    inject(function($injector) {
      basemaps = $injector.get('basemaps');
    });
  });

  it('should create some basemaps', function() {
    expect(basemaps.all).toBeDefined();
  });

  it('should have a valid default basemap', function() {
    expect(basemaps.def).toBeDefined();
    expect(basemaps.all[basemaps.def]).toBeDefined();
  });
});
