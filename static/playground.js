// HTTPTransport is the default transport.
function HTTPTransport() {
	'use strict';

	function playback(output, data) {
		// Backwards compatibility: default values do not affect the output.
		var events = data.Events || [];
		var errors = data.Errors || "";
		var status = data.Status || 0;

		var timeout;
		output({Kind: 'start'});
		function next() {
			if (!events || events.length === 0) {
        if (status > 0) {
          output({Kind: 'end', Body: 'status ' + status + '.'});
        } else {
          if (errors !== "") {
            // errors are displayed only in the case of timeout.
            output({Kind: 'end', Body: errors + '.'});
          } else {
            output({Kind: 'end'});
          }
        }
				return;
			}
			var e = events.shift();
			if (e.Delay === 0) {
				output({Kind: e.Kind, Body: e.Message});
				next();
				return;
			}
			timeout = setTimeout(function() {
				output({Kind: e.Kind, Body: e.Message});
				next();
			}, e.Delay / 1000000);
		}
		next();
		return {
			Stop: function() {
				clearTimeout(timeout);
			}
		};
	}

	function error(output, msg) {
		output({Kind: 'start'});
		output({Kind: 'stderr', Body: msg});
		output({Kind: 'end'});
	}

	function runFailed(output, msg) {
		output({Kind: 'start'});
		output({Kind: 'stderr', Body: msg});
		output({Kind: 'system', Body: '\nShibafu is failed.'});
  }

	var seq = 0;
	return {
		Run: function(body, output) {
			seq++;
			var cur = seq;
			var playing;
			$.ajax('/runtime', {
				type: 'POST',
				data: {"version": 1, "body": body},
				dataType: 'json',
				success: function(data) {
					if (seq != cur) return;
					if (!data) return;
          if (playing != null) playing.Stop();
          if (data.Errors) {
						runFailed(output, data.Errors);
						return;
					}

					playing = playback(output, data);
					return;
				},
				error: function() {
					error(output, 'Error communicating with remote server.');
				}
			});
			return {
				Kill: function() {
					if (playing != null) playing.Stop();
					output({Kind: 'end', Body: 'killed'});
				}
			};
		}
	};
}

function PlaygroundOutput(el) {
	'use strict';

	return function(write) {
		if (write.Kind == 'start') {
			el.innerHTML = '';
			return;
		}

		var cl = 'system';
		if (write.Kind == 'stdout' || write.Kind == 'stderr')
			cl = write.Kind;

		var m = write.Body;
		if (write.Kind == 'end') {
			m = '\nProgram exited' + (m?(': '+m):'.');
		}

		// ^L clears the screen.
		var s = m.split('\x0c');
		if (s.length > 1) {
			el.innerHTML = '';
			m = s.pop();
		}

		m = m.replace(/&/g, '&amp;');
		m = m.replace(/</g, '&lt;');
		m = m.replace(/>/g, '&gt;');

		var needScroll = (el.scrollTop + el.offsetHeight) == el.scrollHeight;

		var span = document.createElement('span');
		span.className = cl;
		span.innerHTML = m;
		el.appendChild(span);

		if (needScroll)
			el.scrollTop = el.scrollHeight - el.offsetHeight;
	};
}

(function() {
  function lineHighlight(error) {
    var regex = /compile:([0-9]+)/g;
    var r = regex.exec(error);
    while (r) {
      $(".lines div").eq(r[1]-1).addClass("lineerror");
      r = regex.exec(error);
    }
  }
  function highlightOutput(wrappedOutput) {
    return function(write) {
      if (write.Body) lineHighlight(write.Body);
      wrappedOutput(write);
    };
  }
  function lineClear() {
    $(".lineerror").removeClass("lineerror");
  }

  // opts is an object with these keys
  //  codeEl - code editor element
  //  outputEl - program output element
  //  runEl - run button element
  //  fmtEl - fmt button element (optional)
  //  fmtImportEl - fmt "imports" checkbox element (optional)
  //  shareEl - share button element (optional)
  //  shareURLEl - share URL text input element (optional)
  //  shareRedirect - base URL to redirect to on share (optional)
  //  toysEl - toys select element (optional)
  //  enableHistory - enable using HTML5 history API (optional)
  //  transport - playground transport to use (default is HTTPTransport)
  //  enableShortcuts - whether to enable shortcuts (Ctrl+S/Cmd+S to save) (default is false)
  //  enableVet - enable running vet and displaying its errors
  function playground(opts) {
    var code = $(opts.codeEl);
    var transport = opts['transport'] || new HTTPTransport(opts['enableVet']);
    var running;

    // autoindent helpers.
    function insertTabs(n) {
      // find the selection start and end
      var start = code[0].selectionStart;
      var end   = code[0].selectionEnd;
      // split the textarea content into two, and insert n tabs
      var v = code[0].value;
      var u = v.substr(0, start);
      for (var i=0; i<n; i++) {
        u += "\t";
      }
      u += v.substr(end);
      // set revised content
      code[0].value = u;
      // reset caret position after inserted tabs
      code[0].selectionStart = start+n;
      code[0].selectionEnd = start+n;
    }
    function autoindent(el) {
      var curpos = el.selectionStart;
      var tabs = 0;
      while (curpos > 0) {
        curpos--;
        if (el.value[curpos] == "\t") {
          tabs++;
        } else if (tabs > 0 || el.value[curpos] == "\n") {
          break;
        }
      }
      setTimeout(function() {
        insertTabs(tabs);
      }, 1);
    }

    // NOTE(cbro): e is a jQuery event, not a DOM event.
    function handleSaveShortcut(e) {
      if (e.isDefaultPrevented()) return false;
      if (!e.metaKey && !e.ctrlKey) return false;
      if (e.key != "S" && e.key != "s") return false;

      e.preventDefault();

      // Share and save
      share(function(url) {
        window.location.href = url + ".w?download=true";
      });

      return true;
    }

    function keyHandler(e) {
      if (opts.enableShortcuts && handleSaveShortcut(e)) return;

      if (e.keyCode == 9 && !e.ctrlKey) { // tab (but not ctrl-tab)
        insertTabs(1);
        e.preventDefault();
        return false;
      }
      if (e.keyCode == 13) { // enter
        if (e.shiftKey) { // +shift
          run();
          e.preventDefault();
          return false;
        } else {
          autoindent(e.target);
        }
      }
      return true;
    }
    code.unbind('keydown').bind('keydown', keyHandler);
    var outdiv = $(opts.outputEl).empty();
    var output = $('<pre/>').appendTo(outdiv);

    function body() {
      return $(opts.codeEl).val();
    }
    function setBody(text) {
      $(opts.codeEl).val(text);
    }
    function origin(href) {
      return (""+href).split("/").slice(0, 3).join("/");
    }

    var pushedEmpty = (window.location.pathname == "/");
    function inputChanged() {
      if (pushedEmpty) {
        return;
      }
      pushedEmpty = true;
      $(opts.shareURLEl).hide();
      window.history.pushState(null, "", "/");
    }
    function popState(e) {
      if (e === null) {
        return;
      }
      if (e && e.state && e.state.code) {
        setBody(e.state.code);
      }
    }
    var rewriteHistory = false;
    if (window.history && window.history.pushState && window.addEventListener && opts.enableHistory) {
      rewriteHistory = true;
      code[0].addEventListener('input', inputChanged);
      window.addEventListener('popstate', popState);
    }

    function setError(error) {
      if (running) running.Kill();
      lineClear();
      lineHighlight(error);
      output.empty().addClass("error").text(error);
    }
    function loading() {
      lineClear();
      if (running) running.Kill();
      output.removeClass("error").text('Waiting for remote server...');
    }
    function run() {
      loading();
      running = transport.Run(body(), highlightOutput(PlaygroundOutput(output[0])));
    }

    var shareURL; // jQuery element to show the shared URL.
    var sharing = false; // true if there is a pending request.
    var shareCallbacks = [];
    function share(opt_callback) {
      if (opt_callback) shareCallbacks.push(opt_callback);

      if (sharing) return;
      sharing = true;

      var sharingData = body();
      $.ajax("/share", {
        processData: false,
        data: sharingData,
        type: "POST",
        contentType: "text/plain; charset=utf-8",
        complete: function(xhr) {
          sharing = false;
          if (xhr.status != 200) {
            alert("Server error; try again.");
            return;
          }
          if (opts.shareRedirect) {
            window.location = opts.shareRedirect + xhr.responseText;
          }
          var path = "/w/" + xhr.responseText;
          var url = origin(window.location) + path;

          for (var i = 0; i < shareCallbacks.length; i++) {
            shareCallbacks[i](url);
          }
          shareCallbacks = [];

          if (shareURL) {
            shareURL.show().val(url).focus().select();

            if (rewriteHistory) {
              var historyData = {"code": sharingData};
              window.history.pushState(historyData, "", path);
              pushedEmpty = false;
            }
          }
        }
      });
    }

    $(opts.runEl).click(run);

    if (opts.shareEl !== null && (opts.shareURLEl !== null || opts.shareRedirect !== null)) {
      if (opts.shareURLEl) {
        shareURL = $(opts.shareURLEl).hide();
      }
      $(opts.shareEl).click(function() {
        share();
      });
    }
  }

  window.playground = playground;
})();
