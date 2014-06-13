(function(){

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
      return diff + ' seconds';
    };
  }])

  // prettyDate
  // ----------
  // turns a unix timestamp into a localized date and time string

  .filter('prettyTime', [function () {
    return function (ts) {
      return new Date(ts * 1000).toLocaleTimeString();
    };
  }])

  // prettyDate
  // ----------
  // turns a unix timestamp into a localized date and time string

  .filter('prettyDate', [function () {
    return function (ts) {
      return new Date(ts * 1000).toLocaleDateString();
    };
  }])

  // prettyDateTime
  // --------------
  // turns a unix timestamp into a localized date and time string

  .filter('prettyDateTime', [function () {
    return function (ts) {
      return new Date(ts * 1000).toLocaleString();
    };
  }])

  // arrLength
  // ---------
  // returns array length

  .filter('arrLength', [function () {
    return function (obj) {
      if (obj && obj.length) {
        return obj.length;
      }
      return 0;
    };
  }])

  // reverse
  // -------
  // returns reversed array

  .filter('reverse', [function() {
    return function(arr) {
      if (!arr || !arr.slice) { return arr; }
      return arr.slice().reverse();
    };
  }])

  // prettyJson
  // ----------
  // takes a string
  // tries to parse it into JSON
  // tries to turn it into formatted JSON (2 space indentation)

  .filter('prettyJson', [function () {
    return function (str) {
      var ret = str;
      try { ret = JSON.parse(ret); } catch (e) {}
      try { ret = JSON.stringify(ret, null, 2); } catch (e) {}
      return ret;
    };
  }]);

})();
