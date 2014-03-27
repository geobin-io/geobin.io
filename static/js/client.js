// (function(){

var map, features, binStore, primus, channel, ws;

var thisUrl = window.location.toString();
var binId = window.location.pathname.substring(1);
var notify = {};
var requests = {};

var navTpl = $('<div class="panel panel-default"> \
  <div class="panel-heading" data-toggle="collapse"><h3 class="panel-title"></h3></div> \
  <div class="list-group collapse"></div></div>');

init();

function init () {
  initMap();
  // initStore();
  bindUI();
  initSocket();
}

// initializer functions

function initStore () {
  if (!geobin.support.localStorage) {
    return;
  }

  binStore = geobin.store.history[binId] = geobin.store.history[binId] || {};
  geobin.save();

  if (!Object.keys(binStore).length) {
    return;
  }

  if ($('.waiting').length) {
    $('.waiting').remove();
  }

  for (var id in binStore) {
    if (binStore.hasOwnProperty(id) && !isNaN(id)) {
      (function(id){
        var d = new Date(+id);
        var reqDate = d.toLocaleDateString();
        var reqTime = d.toLocaleTimeString();
        var reqJson = binStore[id];
        var isNew = !requests[reqDate];
        var navGroup;

        requests[reqDate] = requests[reqDate] || {};

        if (isNew) {
          navGroup = navTpl.clone();
          navGroup.attr('data-date', reqDate);
          navGroup.find('.panel-heading').attr('data-target','.panel[data-date="' + reqDate + '"] .list-group');
          navGroup.find('.panel-title').html('<i class="fa fa-calendar-o"></> ' + reqDate);
        } else {
          navGroup = $('.panel[data-date="' + reqDate + '"]');
        }

        requests[reqDate][reqTime] = reqJson;

        var timestamp = '<i class="fa fa-clock-o"></> ' + reqTime;
        var item = $('<a class="list-group-item" data-id="' + id + '">' + timestamp + '</a>');
        var content = $('<pre class="json" data-id="' + id + '">' + JSON.stringify(reqJson, undefined, 2) + '</pre>');

        searchForGeo(binStore[id]);

        if (isNew) {
          navGroup.hide().prependTo('.callback-nav').fadeIn();
        }

        item.hide().prependTo(navGroup.find('.list-group')).fadeIn();

        content.prependTo('.callback-content');
      })(id);
    }
  }

  var latestDate = $('.callback-nav .list-group').first();
  var latestTime = latestDate.find('>:first-child');
  latestDate.addClass('in');
  latestTime.addClass('active');
  $('.callback-content > [data-id="' + latestTime.data('id') + '"').addClass('active');
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
  ws = new WebSocket(loc + '/' + binId);
  // channel = primus.channel(binId);

  // primus.on('open', function () {
  //   $('.status').text('is listening at ' + thisUrl).fadeIn();

  //   if (!Object.keys(binStore).length) {
  //     notify.success();
  //   }
  // });

  // channel.on('data', processData);
  ws.onmessage = function () {
    console.log(arguments);
  };
}

// helper functions

notify.success = function () {
  var alerts = $('.alerts');
  var alert = $('<div class="alert alert-success"></div>');

  var sampleJson = '{"geo":{"latitude":"45.5165","longitude":"-122.6764"}}';

  var code = 'curl -X POST -H "Content-Type: application/json" -d \'' + sampleJson + '\' ' + thisUrl;

  alerts.empty();

  alert
    .html('<strong>Connected!</strong> Try running <code>' + code + '</code> to get some feedback.')
    .hide()
    .prependTo('.alerts')
    .fadeIn();
};

function processData (data) {
  $('.alerts').empty();

  if ($('.waiting').length) {
    $('.waiting').remove();
  }

  var id = new Date().getTime();
  var d = new Date(id);
  var reqDate = d.toLocaleDateString();
  var reqTime = d.toLocaleTimeString();
  var reqJson = data;
  var isNew = !requests[reqDate];
  var navGroup;

  requests[reqDate] = requests[reqDate] || {};

  if (geobin.support.localStorage) {
    geobin.store.history[binId][id] = reqJson;
    geobin.save();
  }

  if (isNew) {
    navGroup = navTpl.clone();
    navGroup.attr('data-date', reqDate);
    navGroup.find('.panel-heading').attr('data-target','.panel[data-date="' + reqDate + '"] .list-group');
    navGroup.find('.panel-title').html('<i class="fa fa-calendar-o"></> ' + reqDate);
  } else {
    navGroup = $('.panel[data-date="' + reqDate + '"]');
  }

  requests[reqDate][reqTime] = reqJson;

  try {
    data = JSON.stringify(data, undefined, 2);
  } catch (e) {}

  searchForGeo(JSON.parse(data));

  var timestamp = '<i class="fa fa-clock-o"></> ' + reqTime;
  var item = $('<a class="list-group-item" data-id="' + id + '">' + timestamp + '</a>');
  var content = $('<pre class="json" data-id="' + id + '">' + data + '</pre>');

  if (isNew) {
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

//})();