(function(){

  'use strict';

  // Filters
  angular.module('Geobin.filters', [])

  // timeRemaining
  // -------------
  // takes a unix timestamp
  // returns a humanized number in hours, minutes, or seconds

  .filter('timeRemaining', [function () {
    return function (ts) {
      var n = Math.floor(new Date().getTime() / 1000);
      var diff = ts - n;

      // bin is expired
      if (diff < 1) {
        return 'expired';
      }

      // bin has more than an hour left
      if (diff > 3600) {
        return Math.floor(diff / 3600) + ' hours';
      }

      // bin has more than a minute left
      if (diff > 60) {
        return Math.floor(diff / 60) + ' minutes';
      }

      // default to seconds
      return diff + 'seconds';
    };
  }])

  // prettyDate
  // ----------

  .filter('prettyDate', [function () {
    return function (ts) {
      return new Date(ts * 1000).toLocaleString();
    };
  }])

  // objLength
  // ----------

  .filter('length', [function () {
    return function (obj) {
      if (typeof obj === 'array') {
        return obj.length;
      }
      return '';
    };
  }])

  // jsonStringify
  // ----------

  .filter('jsonStringify', [function () {
    return function (str) {
      var ret = str;
      try {
        var ret = JSON.parse(ret);
      } catch (e) {}
      try {
        var ret = JSON.stringify(ret, null, 2);
      } catch (e) {}
      return ret;
    };
  }]);

})();
