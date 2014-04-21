(function(){

  'use strict';

  // Filters
  angular.module('Geobin.filters', [])

  // Interpolate filter
  .filter('interpolate', ['version', function(version) {
    return function(text) {
      return String(text).replace(/\%VERSION\%/mg, version);
    };
  }])

  .filter('prettyDate', [function(){
    return function(ts) {
      var d = new Date(ts * 1000);
      var dateString = d.getHours() + ':' + d.getMinutes() + ' ' +
        + ('0' + (d.getMonth()+1)).slice(-2) + '/'
        + ('0' + d.getDate()).slice(-2) + '/'
        + d.getFullYear();
      return dateString;
    }
  }]);

})();
