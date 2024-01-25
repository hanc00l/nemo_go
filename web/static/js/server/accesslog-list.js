$(function () {
    //$('#btnsiderbar').click();
    load_accesslog_file_list();
    //搜索
    $("#search").click(function () {
        load_accesslog();
    });

});


function load_accesslog_file_list() {
    $.post("/accesslog-file", {}, function (data, e) {
        if (e === "success") {
            for (let i = 0; i < data.length; i++) {
                $("#log_file_num").append("<option value='" + i + "'>" + data[i] + "</option>");
            }
        }
    });
}


function load_accesslog() {
    $.post("/accesslog-list",
        {
            "file_num": $('#log_file_num').val(),
        }, function (data, e) {
            if (e === "success" && data['status'] === "success") {
                $('#text_accesslog').val(data['msg']);
            } else {
                $('#text_accesslog').val('load accesslog file fail!');
            }
        });
}