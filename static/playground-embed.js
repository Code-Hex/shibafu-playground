// opts is an object with these keys
//  codeEl - code editor element
//  embedEl - embed checkbox element
//  embedLabelEl - embed label element, containing embedEl
//  embedHTMLEl - embed HTML text input element
function playgroundEmbed(opts) {
  if (opts.codeEl === null || opts.embedEl === null || opts.embedLabelEl === null || opts.embedHTMLEl === null) {
    return;
  }

  var code = $(opts.codeEl);
  var embed = $(opts.embedEl);
  var embedLabel = $(opts.embedLabelEl);

  function inIFrame(){
    return window.self !== window.top;
  }
  embedLabel.hide();
  if (inIFrame()) {
    $("body").addClass("embedded");
    return;
  }

  function origin(href) {
    return (""+href).split("/").slice(0, 3).join("/");
  }

  function inputChanged() {
    embedLabel.hide();
  }

  if (window.history && window.history.pushState && window.addEventListener) {
    code[0].addEventListener('input', inputChanged);
  }

  var embedHTML = $(opts.embedHTMLEl).hide();
  var embedding = false;
  embed.change(function() {
    if (embedding) return;
    embedding = true;
    var embeddingData = code.val();
    $.ajax("/share", {
      processData: false,
      data: embeddingData,
      type: "POST",
      complete: function(xhr) {
        embedding = false;
        if (xhr.status != 200) {
          alert("Server error; try again.");
          return;
        }
        if (embedHTML) {
          var path = "/w/" + xhr.responseText;
          var url = origin(window.location) + path;
          if (embed.prop('checked')){
            url = "<iframe src=\"" + url + "\" frameborder=\"0\" style=\"width: 100%; height: 100%\"><a href=\"" + url + "\">see this code in play.shibafu.org</a></iframe>";
          }
          embedHTML.show().val(url).focus().select();
        }
      }
    });
  });
}
