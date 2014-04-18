'use strict';

// Directives
angular.module('Geobin.directives', [])

// App Version
.directive('appVersion', ['version', function(version) {
  return function(scope, elm, attrs) {
    elm.text(version);
  };
}]);
