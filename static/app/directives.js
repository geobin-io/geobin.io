angular.module('Geobin.directives')

// App Version
.directive('clientVersion', ['clientVersion', function (version) {
  return function (scope, elm, attrs) {
    elm.text(version);
  };
}])

// API Version
.directive('apiVersion', ['apiVersion', function (version) {
  return function (scope, elm, attrs) {
    elm.text(version);
  };
}]);
