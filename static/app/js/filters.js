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

  .filter('timeRemaining', [function(){
    return function(ts) {
      var n = Math.floor(new Date().getTime() / 1000);
      var diff = ts - n;
      if (diff < 1) {
        return 'expired';
      }
      if (diff > 3600) {
        return Math.floor(diff / 3600) + ' hours';
      }
      if (diff > 60) {
        return Math.floor(diff / 60) + ' minutes';
      }
      return diff + 'seconds';
    };
  }]);

})();
