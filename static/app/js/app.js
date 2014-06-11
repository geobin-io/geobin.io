(function(){

  angular.module('Geobin', [
    'ui.router',
    'Geobin.filters',
    'Geobin.services',
    'Geobin.directives',
    'Geobin.controllers',
    'chieffancypants.loadingBar',
    'yaru22.md'
  ])

  .config(['$locationProvider', '$stateProvider', '$urlRouterProvider',
    function ($locationProvider, $stateProvider, $urlRouterProvider) {
      $locationProvider.html5Mode(true);

      $urlRouterProvider.when(/^\/api.*/, '/doc/api');
      $urlRouterProvider.otherwise('/');

      $stateProvider
       .state('home', {
         url: '/',
         controller: 'HomeCtrl',
         templateUrl: '/static/app/partials/home.html'
       })
       .state('api', {
         url: '/doc/api',
         templateUrl: '/static/app/partials/doc/api.html'
       })
       .state('client', {
         url: '/doc/client',
         templateUrl: '/static/app/partials/doc/client.html'
       })
       .state('server', {
         url: '/doc/server',
         templateUrl: '/static/app/partials/doc/server.html'
       })
       .state('bin', {
         url: '/:binId',
         controller: 'BinCtrl',
         templateUrl: '/static/app/partials/bin.html'
       });
    }
  ]);

})();
