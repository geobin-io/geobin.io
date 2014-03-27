(function(){

var $history = $('.history-nav');

if (!geobin.support.localStorage) {
  return $history.append('<li class="list-group-item">not supported by your browser</li>');
}

var history = geobin.store.history;

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

$('.clear-history').click(function(e){
  e.preventDefault();
  geobin.clearHistory();
  $history.html('<li class="list-group-item">none yet</li>');
});

})();