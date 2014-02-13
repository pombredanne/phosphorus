$('#uploadButton').click(function() {
    var files = $("#filePicker")[0].files;
    var lastFn = function() {};
    $.each(files, function(index) {
	file = files[index];
	f = upload(file, lastFn, index);
	lastFn = f;
    });
    lastFn();
});

function formData(data) {
    fd = new FormData();
    fd.append('key', data.key);
    fd.append('AWSAccessKeyId', data.AWSAccessKeyId);
    fd.append('acl', 'private');
    fd.append('success_action_status', '200');
    fd.append('Content-Type', "text/csv");
    fd.append('policy', data.policy);
    fd.append('signature', data.signature);
    return fd
}

function upload(file, cb, index) {
    var fname = file.name
    return function() {
	$.ajax({
	    url: "/_form",
	    type: 'GET',
	    dataType: "json",
	    success: function(data) {
		fd = formData(data);
		fd.append('file', file);
		$("#progressBars").append('<div><progress id="progress_' + index + '" value="0" max="100"></progress>' + fname + '</div>');
		$.ajax({
		    url: 'http://phosphorus-upload.s3.amazonaws.com/',
		    type: 'POST',
		    xhr: function() {
			var myXhr = $.ajaxSettings.xhr();
			if(myXhr.upload){
			    myXhr.upload.addEventListener('progress', function(e) {
				if(e.lengthComputable){
				    $('#progress_' + index).attr({value:e.loaded,max:e.total});
				}
			    }, false);
			}
			return myXhr;
		    },
		    success: function(e) {
			$('#progress_' + index).after("<b>OK</b>");
			cb();
		    },
		    data: fd,
		    cache: false,
		    contentType: false,
		    processData: false
		});
	    }
	});
    };
}
