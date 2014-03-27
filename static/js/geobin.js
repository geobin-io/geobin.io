(function(global){

  var geobin = {};
  geobin.store = null;
  geobin.support = {};

  geobin.support.localStorage = (function(){
    try {
      return 'localStorage' in window && window['localStorage'] !== null;
    } catch (e) {
      return false;
    }
  })();

  if (geobin.support.localStorage) {
    try {
      geobin.store = JSON.parse(localStorage.getItem('geobin')) || {};
    } catch (e) {
      geobin.store = {};
    }
    geobin.store.history = geobin.store.history || {};
    geobin.save = function () {
      localStorage.setItem('geobin', JSON.stringify(geobin.store));
    }
    geobin.clearHistory = function () {
      localStorage.setItem('geobin', {});
    }
  }

  global.geobin = geobin;

})(this);
