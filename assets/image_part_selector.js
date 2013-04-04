(function(global) {
  function ImagePartSelector(element) {
    if (!element || element.tagName !== 'CANVAS') {
      throw new Error('Please provide a valid canvas element');
    }

    this.element = element;
    this.context = this.element.getContext('2d');

    this._startX = 0;
    this._startY = 0;

    this._rect   = this.element.getBoundingClientRect();
    this._image;

    this.bindEvents();
  };

  ImagePartSelector.prototype.bindEvents = function() {
    var self      = this;
    var mousemove = function(event) {
      var width  = (event.clientX - self._rect.top) - self._startX;
      var height = (event.clientY - self._rect.left) - self._startY;

      self.clearCanvas();
      self.drawCurrentImage();

      self.context.rect(self._startX, self._startY, width, height);
      self.context.lineWidth   = 1;
      self.context.strokeStyle = 'black';
      self.context.stroke();
    };

    this.element.addEventListener('mousedown', function(event) {
      self._startX = event.clientX - self._rect.left;
      self._startY = event.clientY - self._rect.top;
      self.element.addEventListener('mousemove', mousemove);
    });

    this.element.addEventListener('mouseup', function(event) {
      self.element.removeEventListener('mousemove', mousemove);
    });
  };

  ImagePartSelector.prototype.loadImage = function(source) {
    source             = source || 'http://coding-robin.de/images/pictures/spotify-remote.png';
    this._image        = new Image();
    this._image.onload = this.drawCurrentImage.bind(this);
    this._image.src    = source;
  };

  ImagePartSelector.prototype.drawCurrentImage = function() {
    this.element.width  = this._image.width;
    this.element.height = this._image.height;
    this.context.drawImage(this._image, 0, 0, this._image.width, this._image.height);
  };

  ImagePartSelector.prototype.clearCanvas = function() {
    this.context.clearRect(0, 0, this.element.width, this.element.height);
  };

  global.ImagePartSelector = ImagePartSelector;
})(window);
