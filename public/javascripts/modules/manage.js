'use strict';

angular.module('ManageModule', ['ngUpload'])
.controller('ManageCtrl', function ($scope, $rootScope, $http) {
  $scope.newFeedUrl = null;
  $rootScope.currentPage = 'manage';
  $scope.predicate = 'title';

  $scope.refresh = function() {
    $http.get('/api/subscription').success(function (data) {
      $scope.subscriptions = data;
    });
  };

  $scope.getStatusClass = function(feed) {
    var result = 'default';
    if (feed.status) {
      if (feed.status.match(/^error/g)) result = 'danger';
      else if (feed.pshbStatus == 'subscribe') result = 'primary';
      else result = 'success';
    }
    return result;
  };

  $scope.getStatusDesc = function(feed) {
    var result = 'This feed had not yet been parsed.';
    if (feed.status) {
      if (feed.status.match(/^error/g)) result = feed.status;
      else if (feed.pshbStatus == 'subscribe') result = 'Feed subscribed to ' + feed.hub;
      else result = 'Feed successfully parsed';
    }
    return result;
  };

  $scope.getStatusLabel = function(status) {
    var result = 'Not updated';
    if (status) {
      result = status.match(/^error/g) ? 'In error' : status;
    }
    return result;
  };

  $scope.getHubUrl = function(feed) {
    var path = '/subscription-details?';
    var query = {
      'hub.callback': window.location.origin + '/pubsubhubbud/callback',
      'hub.topic': feed.xmlurl
    };
    var result = 'Not updated';
    if (status) {
      result = status.match(/^error/g) ? 'In error' : status;
    }
    return feed.hub + path + $.param(query);
  };

  $scope.unSubscribe = function(feed) {
    if (confirm('Do you really want to unsubscribe from "' + feed.title + '" ?')) {
      $http.delete('/api/subscription/' + feed.id)
      .success(function(data) {
        $scope.message = {clazz: 'alert-success', text: 'Feed "' + feed.title + '" successfully removed.'};
        for (var i = 0; i < $scope.subscriptions.length; i++) {
          if ($scope.subscriptions[i].id == feed.id) {
            $scope.subscriptions.splice(i, 1);
            $scope.newFeedUrl = feed.xmlurl;
            $rootScope.$broadcast('app.event.subscriptions.remove', feed);
            break;
          }
        }
      })
      .error(function() {
        $scope.message = {clazz: 'alert-danger', text: 'Unable to remove Feed "' + feed.title + '"!'};
      });
    }
  };

  $scope.subscribe = function(url) {
    if (url) {
      $http.post('/api/subscription', {url: url})
      .success(function(feed) {
        $scope.message = {clazz: 'alert-success', text: 'Feed "' + feed.title + '" successfully added.'};
        $scope.subscriptions.push(feed);
        $scope.newFeedUrl = null;
        $rootScope.$broadcast('app.event.subscriptions.add', feed);
      })
      .error(function() {
        $scope.message = {clazz: 'alert-danger', text: 'Unable to remove Feed "' + url + '"!'};
      });
    }
  };

  $scope.uploadComplete = function (content, completed) {
    if (completed) {
      var res = {};
      if (content.error) {
        res.error = content.error;
      }
      if (res.error) {
        $scope.message = {clazz: 'alert-danger', text: 'Unable to import file: ' + res.error};
      } else {
        $scope.message = {clazz: 'alert-success', text: content.length + ' subscription(s) successfully imported.'};
        $scope.refresh();
        $rootScope.$broadcast('app.event.subscriptions.refresh');
      }
    }
  };

  $scope.refresh();
});
