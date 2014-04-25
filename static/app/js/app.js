(function(){

  angular.module('Geobin', [
    'ngRoute',
    'Geobin.filters',
    'Geobin.services',
    'Geobin.directives',
    'Geobin.controllers',
    'chieffancypants.loadingBar'
  ])

  .config(['$routeProvider', '$locationProvider', function ($routeProvider, $locationProvider) {
    $locationProvider.html5Mode(true);

    $routeProvider
      .when('/', {
        controller: 'HomeCtrl',
        templateUrl: '/static/app/partials/home.html'
      })
      .when('/:binId', {
        controller: 'BinCtrl',
        templateUrl: '/static/app/partials/bin.html'
      })
      .otherwise({redirectTo: '/'});
  }]);

})();
