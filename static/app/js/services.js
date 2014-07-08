(function(){

  angular.module('Geobin.services', [])

  /**
   * clientVersion
   * -------------
   * client version constant
   */
  .value('clientVersion', '1.0.3')

  /**
   * apiVersion
   * ----------
   * API version constant
   */
  .value('apiVersion', '1')

  /**
   * basemaps
   * --------
   * houses all available basemaps, specifies default
   */
  .service('basemaps', function () {
    this.init = function () {
      var streets = L.esri.basemapLayer('Streets');
      var topo = L.esri.basemapLayer('Topographic');
      var oceans = L.esri.basemapLayer('Oceans');
      var natgeo = L.esri.basemapLayer('NationalGeographic');
      var gray = L.layerGroup([
        L.esri.basemapLayer('Gray'),
        L.esri.basemapLayer('GrayLabels')
      ]);
      var darkgray = L.layerGroup([
        L.esri.basemapLayer('DarkGray'),
        L.esri.basemapLayer('DarkGrayLabels')
      ]);
      var imagery = L.layerGroup([
        L.esri.basemapLayer('Imagery'),
        L.esri.basemapLayer('ImageryLabels')
      ]);
      var shadedrelief = L.layerGroup([
        L.esri.basemapLayer('ShadedRelief'),
        L.esri.basemapLayer('ShadedReliefLabels')
      ]);

      this.all = {
        'Streets': streets,
        'Topographic': topo,
        'Oceans': oceans,
        'Nat Geo': natgeo,
        'Gray': gray,
        'Dark Gray': darkgray,
        'Imagery': imagery,
        'Shaded Relief': shadedrelief
      };

      this.def = 'Streets';
    };

    this.init();
  })

  /**
   * store
   * -----
   * localStorage interface for browser-based user persistence
   */
  .service('store', function () {

    var local = new TinyStore('geobin');

    (function cleanHistory () {
      var h = local.session.history = local.session.history || [];
      var n = Math.floor(new Date().getTime() / 1000);

      for (var i = h.length - 1; i >= 0; i--) {
        var diff = h[i].expires - n;
        if (diff < 1) {
          h.splice(i, 1);
        }
      }

      local.save();
    })();

    this.local = local;

  })

  /**
   * api
   * ---
   * service for interacting with the Geobin API
   */

  .service('api', ['$http', '$location', 'store', 'apiVersion', function ($http, $location, store, apiVersion) {

    // create reference to self for use in child methods
    var api = this;

    /**
     * endpoint
     * --------
     * expose base API endpoint
     * @type {String}
     */
    this.endpoint = '/api/' + apiVersion + '/';

    /**
     * create
     * ------
     * create a new geobin
     */
    this.create = function () {
      var route = api.endpoint + 'create';

      $http.post(route, {})
        .success(function createSuccess (data, status, headers, config) {
          store.local.session.history.push(data);
          store.local.save();
          $location.path('/' + data.id);
        });
    };

    /**
     * history
     * -------
     * POST to /api/{apiVersion}/history/{binId}
     * get bin history object by ID
     * @param  {String}   binId     ID of bin
     * @param  {Function} callback  function to call on successful POST
     */
    this.history = function (binId, callback) {
      var route = api.endpoint + 'history/' + binId;

      $http.post(route, {})
        .success(function historySuccess (data, status, headers, config) {
          if (status === 200) {
            callback(data);
          }
        })
        .error(function historyError (data, status, headers, config) {
          if (status === 404) {
            callback();
          }
        });
    };

    /**
     * counts
     * ------
     * get number of requests made to bins by array of bin IDs
     * @param  {Array}    binIds    Array of bin IDs
     * @param  {Function} callback  function to call on successful POST
     */
    this.counts = function (binIds, callback) {
      var route = api.endpoint + 'counts';

      $http.post(route, binIds)
        .success(function countsSuccess (data, status, headers, config) {
          if (status === 200) {
            callback(data);
          }
        });
    };

    /**
     * Sockets
     * -------
     * Methods pertaining to opening and closing WebSocket connections for a bin
     */

    this.ws = {};

    /**
     * sockets
     * -------
     * Collection of WebSockets by bin ID
     */
    var sockets = this.ws.sockets = {};

    /**
     * ws.open
     * -------
     * Creates a new WebSocket connection and binds a callback to `onmessage`
     * @param  {String}   binId     ID of bin to connect to
     * @param  {Function} callback  function to bind to `onmessage`
     */
    this.ws.open = function (binId, callback) {
      var route = 'ws://' + window.location.host + api.endpoint + 'ws/' + binId;

      sockets[binId] = new WebSocket(route);
      sockets[binId].onmessage = callback;
    };

    /**
     * ws.close
     * --------
     * Closes a WebSocket connection
     * @param  {String} binId   ID of bin to close
     */
    this.ws.close = function (binId) {
      if (sockets[binId] && sockets[binId].close) {
        sockets[binId].close();
      }
    };

  }]);

})();
