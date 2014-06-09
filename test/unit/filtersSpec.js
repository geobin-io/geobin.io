describe('filters', function() {
  beforeEach(module('Geobin.filters'));

  describe('timeRemaining', function() {
    it('should take a unix timestamp 1 second in the future and return the number of seconds it is from now', inject(function(timeRemainingFilter) {
      var d = Math.floor(new Date().getTime()/1000) + 1;
      expect(timeRemainingFilter(d)).toBe('1 seconds');
      expect(timeRemainingFilter(d+5)).toBe('6 seconds');
    }));

    it('should take a unix timestamp 1 minute in the future and return the number of minutes it is from now', inject(function(timeRemainingFilter) {
      var d = Math.floor(new Date().getTime()/1000) + 61;
      expect(timeRemainingFilter(d)).toBe('1 minutes');
      expect(timeRemainingFilter(d+60)).toBe('2 minutes');
    }));

    it('should take a unix timestamp 1 hour in the future and return the number of hours it is from now', inject(function(timeRemainingFilter) {
      var d = Math.floor(new Date().getTime()/1000) + 3601;
      expect(timeRemainingFilter(d)).toBe('1 hours');
      expect(timeRemainingFilter(d+3600)).toBe('2 hours');
    }));

    it('should take a unix timestamp 1 hour in the future and return the number of hours it is from now', inject(function(timeRemainingFilter) {
      var d = Math.floor(new Date().getTime()/1000) + 3601;
      expect(timeRemainingFilter(d)).toBe('1 hours');
    }));

    it('should take a unix timestamp from the past and return "expired"', inject(function(timeRemainingFilter) {
      var d = Math.floor(new Date().getTime()/1000) - 1;
      expect(timeRemainingFilter(d)).toBe('expired');
    }));
  });

  describe('prettyDate', function() {
    it('should turn a unix timestamp into a localized date/time string', inject(function(prettyDateFilter) {
      var d = new Date();
      var ts = Math.floor(d.getTime()/1000);
      expect(prettyDateFilter(ts)).toBe(d.toLocaleString());
    }));
  });

  describe('arrLength', function() {
    it('should return the proper length of a given array', inject(function(arrLengthFilter) {
      expect(arrLengthFilter([0,1,2])).toBe(3);
      expect(arrLengthFilter([])).toBe(0);
    }));

    it('should return 0 for inputs that are not an array', inject(function(arrLengthFilter) {
      expect(arrLengthFilter(0)).toBe(0);
      expect(arrLengthFilter()).toBe(0);
    }));
  });

  describe('reverse', function() {
    it('should reverse the items in an array', inject(function(reverseFilter) {
      expect(reverseFilter([0,1,2])).toEqual([2,1,0]);
    }));
  });

  describe('prettyJson', function() {
    it('should make ugly json pretty', inject(function(prettyJsonFilter) {
      expect(prettyJsonFilter('{"this":"is some pretty ugly json",        "ami":  \n    "rite?"\n}\t')).toBe('{\n  "this": "is some pretty ugly json",\n  "ami": "rite?"\n}');
    }));
  });
});
