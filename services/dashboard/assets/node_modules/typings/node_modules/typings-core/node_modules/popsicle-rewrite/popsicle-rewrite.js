module.exports = rewritePopsicle

function rewritePopsicle (rewrites) {
  var mappings = Object.keys(rewrites).map(function (pattern) {
    return [ new RegExp(pattern), rewrites[pattern] ]
  })

  return function (request, next) {
    for (var i = 0; i < mappings.length; i++) {
      var mapping = mappings[i]
      if (request.url.search(mapping[0]) >= 0) {
        request.url = request.url.replace(mapping[0], mapping[1])
        break
      }
    }
    return next()
  }
}
