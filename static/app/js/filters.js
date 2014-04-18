'use strict';

// Filters
angular.module('Geobin.filters', [])

// Interpolate filter
.filter('interpolate', ['version', function(version) {
  return function(text) {
    return String(text).replace(/\%VERSION\%/mg, version);
  };
}]);
