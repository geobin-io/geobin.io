(function(){

  angular.module('Geobin', [
    'ui.router',
    'Geobin.filters',
    'Geobin.services',
    'Geobin.directives',
    'Geobin.controllers',
    'chieffancypants.loadingBar'
  ])

  .config(['$locationProvider', '$stateProvider', '$urlRouterProvider',
    function ($locationProvider, $stateProvider, $urlRouterProvider) {
      $locationProvider.html5Mode(true);

      $urlRouterProvider.otherwise('/');

      $stateProvider
       .state('home', {
         url: '/',
         controller: 'HomeCtrl',
         templateUrl: '/static/app/partials/home.html'
       })
       .state('bin', {
         url: '/:binId',
         controller: 'BinCtrl',
         templateUrl: '/static/app/partials/bin.html'
       });
    }
  ]);

})();
