(function(global){

  // SUPPORT
  // -------
  // object to test and store support for required things
  // namely localStorage and WebSocket

  var support = {};

  support.localStorage = (function(){
    try {
      return 'localStorage' in window && window['localStorage'] !== null;
    } catch (e) {
      return false;
    }
  })();

  support.WebSocket = (function(){
    try {
      return 'WebSocket' in window && window['WebSocket'] !== null;
    } catch (e) {
      return false;
    }
  })();

  // STORE
  // -----
  // object to interact with geobin history through localStorage

  var store = null;

  if (support.localStorage) {
    try {
      store = JSON.parse(localStorage.getItem('geobin')) || {};
    } catch (e) {
      store = {};
    }

    store.history = store.history || {};

    store.save = function () {
      localStorage.setItem('geobin', JSON.stringify(store.history));
    };

    store.clearHistory = function () {
      localStorage.setItem('geobin', {});
    };
  }

  // GEOBIN
  // ------
  // Constructor!
  // Figures out if you're in a bin or in the index then delegates work.

  function Geobin () {
    this.url = window.location.toString();
    this.binId = window.location.pathname.substring(1);
    this.support = support;
    this.store = store;

    // ROUTER (lol)
    if (this.binId !== '') {
      bin.apply(this);
    } else if (support.localStorage) {
      index.apply(this);
    }
  }

  // BIN
  // ---
  // Logic for when you're in a bin!
  // Requests bin history, sets up socket cxn & map, displays requests
  // Updates localStorage, processes request data
  // TODO: clean up this mess

  function bin () {
    var map, features;

    var notify = {};
    var requests = {};
    var binId = this.binId;
    var url = this.url;

    var navTpl = $('<div class="panel panel-default">' +
      '<div class="panel-heading" data-toggle="collapse"><h3 class="panel-title"></h3></div>' +
      '<div class="list-group collapse"></div></div>');

    function isJson(str) {
      try {
        JSON.parse(str);
      } catch (e) {
        return false;
      }
      return true;
    }

    getHistory();
    initMap();
    bindUI();
    initSocket();

    // initializer functions

    function getHistory () {
      $.get('/api/history/' + binId, function(data) {
        console.log(data);
      });
    }

    function initMap () {
      map = L.map('map', {
        center: [0,0],
        zoom: 1,
        scrollWheelZoom: true,
        zoomControl: false
      });

      features = new L.FeatureGroup();

      map.addLayer(features);

      new L.Control.Zoom({ position: 'topright' }).addTo(map);

      L.esri.basemapLayer('Streets').addTo(map);
    }

    function bindUI () {
      $('.callback-nav').on('click', 'a[data-id]', function(e){
        e.preventDefault();

        var id = $(this).data('id');
        $('.active').removeClass('active');
        $('[data-id="' + id + '"]').addClass('active');
      });
    }

    function initSocket () {
      var loc = window.location.origin.replace(/https?/,'ws');
      var ws = new WebSocket(loc + '/api/ws/' + binId);

      ws.onopen = function () {
        $('.status').text('is listening at ' + url).fadeIn();
        notify.success();
      };

      ws.onmessage = function (event) {
        console.log(event);
        if (event && event.data && isJson(event.data)) {
          console.log('hi');
          processData(JSON.parse(event.data));
        }
      };
    }

    // helper functions

    notify.success = function () {
      var alerts = $('.alerts');
      var alert = $('<div class="alert alert-success"></div>');
      var sampleJson = '{"geo":{"latitude":"45.5165","longitude":"-122.6764"}}';
      var code = 'curl -X POST -H "Content-Type: application/json" -d \'' + sampleJson + '\' ' + this.url;
      var html = '<strong>Connected!</strong> Try running <code>' + code + '</code> to get some feedback.';

      alerts.empty().prepend(alert.html(html));
    };

    function processData (data) {
      $('.alerts').empty();

      if ($('.waiting').length) {
        $('.waiting').remove();
      }

      var id = Number(data.timestamp);
      var d = new Date(id);
      var reqDate = d.toLocaleDateString();
      var reqTime = d.toLocaleTimeString();
      var body = data.body;
      var isNewDate = !requests[reqDate];
      var navGroup;

      requests[reqDate] = requests[reqDate] || {};

      if (isNewDate) {
        navGroup = navTpl.clone();
        navGroup.attr('data-date', reqDate);
        navGroup.find('.panel-heading').attr('data-target','.panel[data-date="' + reqDate + '"] .list-group');
        navGroup.find('.panel-title').html('<i class="fa fa-calendar-o"></> ' + reqDate);
      } else {
        navGroup = $('.panel[data-date="' + reqDate + '"]');
      }

      requests[reqDate][reqTime] = body;

      if (isJson(body)) {
        body = JSON.parse(body);
        searchForGeo(body);
      }

      var timestamp = '<i class="fa fa-clock-o"></> ' + reqTime;
      var item = $('<a class="list-group-item" data-id="' + id + '">' + timestamp + '</a>');
      var content = $('<pre class="json" data-id="' + id + '">' + syntaxHighlight( JSON.stringify(body, undefined, 2) ) + '</pre>');

      if (isNewDate) {
        navGroup.hide().prependTo('.callback-nav').fadeIn();
      }

      item.hide().prependTo(navGroup.find('.list-group')).fadeIn();
      content.prependTo('.callback-content');
    }

    function syntaxHighlight (json) {
      json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
      return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
          if (/:$/.test(match)) {
            cls = 'key';
          } else {
            cls = 'string';
          }
        } else if (/true|false/.test(match)) {
          cls = 'boolean';
        } else if (/null/.test(match)) {
          cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
      });
    }

    function searchForGeo (data) {
      if (Object.prototype.toString.call(data) !== '[object Object]') {
        return;
      }
      for (var key in data) {
        if (data.hasOwnProperty(key)) {
          var obj = data[key];
          if (key === 'geo') {
            mapGeo(obj);
          } else {
            // DANGER: RECURSION
            searchForGeo(obj);
          }
        }
      }
    }

    // mapping according to geo object spec for now
    // https://developers.arcgis.com/en/geotrigger-service/api-reference/geo-objects/
    // only doing point, point+radius, & geojson (polygon/multipolygon)
    function mapGeo (geo) {
      var layer;

      if (geo.latitude && geo.longitude && geo.distance) {
        layer = new L.Circle([geo.latitude, geo.longitude], geo.distance);
      }
      else if (geo.latitude && geo.longitude) {
        layer = new L.Marker([geo.latitude, geo.longitude]);
      }
      else if (geo.geojson) {
        layer = new L.GeoJSON(geo.geojson);
      }

      if (layer) {
        layer.addTo(features);

        if (geo.distance || geo.geojson) {
          var bounds = layer.getBounds();
          map.fitBounds(bounds);
        }
      }
    }
  }

  // INDEX
  // -----
  // Logic for when you're at the index!
  // Populates history menu, then just chills

  function index () {
    console.log('index');

    var history  = store.history,
        $history = $('.history'),
        $nav     = $history.find('.history-nav'),
        $clear   = $('.clear-history');

    if (Object.keys(history).length) {
      for (var item in history) {
        if (history.hasOwnProperty(item)) {
          var length = Object.keys(history[item]).length;
          var content = item;
          if (length) {
            content += '<span class="badge pull-right">' + length + '</span>';
          }
          $history.append('<li class="list-group-item"><a href="' + item + '">' + content + '</a></li>');
        }
      }
    } else {
      $history.append('<li class="list-group-item">none yet</li>');
    }

    $history.removeClass('hide');

    $clear.click(function(e){
      e.preventDefault();
      store.clearHistory();
      $history.html('<li class="list-group-item">none yet</li>');
    });
  }

  // INSTANTIATION
  // -------------

  global.geobin = new Geobin();

})(this);
