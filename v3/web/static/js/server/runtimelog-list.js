$(function () {
    $.urlParam = function (param) {
        const results = new RegExp(`[?&]${param}=([^&#]*)`).exec(window.location.href);
        return results ? results[1] : null;
    };
    $('#runtimelog_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/runtimelog-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "log_source": $('#log_source').val(),
                        "log_func": $('#log_func').val(),
                        "log_level": $('#log_level').val(),
                        "log_message": $('#log_message').val(),
                        "date_delta": $('#date_delta').val()
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    width: "5%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        const strData = '<input type="checkbox" class="checkchild" value="' + row['id'] + '"/>';
                        return strData;
                    }
                },
                {
                    data: "index", title: "序号", width: "5%"
                },
                {
                    data: "source", title: "Source", width: "15%",
                    render: function (data, type, row, meta) {
                        let strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                        strData += data;
                        strData += '</div>'
                        return strData;
                    }
                },
                {
                    data: "func", title: "Func", width: "15%",
                    "render": function (data, type, row, meta) {
                        let strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                        strData += '<a href="/runtimelog-info?id=' + row['id'] + '" target="_blank">' + data + '</a>';
                        strData += '</div>'
                        return strData;
                    }
                },
                {
                    data: 'message', title: 'Message', width: '30%',
                    "render": function (data, type, row, meta) {
                        let strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                        strData += encodeHtml(data);
                        strData += '</div>'
                        return strData;
                    }
                },
                {
                    data: 'level', title: 'Level', width: '5%',
                    render: function (data, type, row, meta) {
                        if (data === "warning") {
                            return '<span class="text-warning">' + data + '</span>';
                        } else if (data === "error" || data === "fatal") {
                            return '<span class="text-danger">' + data + '</span>';
                        } else return data;
                    }
                },
                {
                    data: 'update_datetime', title: '时间', width: '15%'
                },
                {
                    title: "操作",
                    width: "8%",
                    "render": function (data, type, row, meta) {
                        const strDelete = "<a class=\"btn btn-sm btn-danger\" href=javascript:delete_runtimelog(\"" + row["id"] + "\") role=\"button\" title=\"Delete\"><i class=\"fa fa-trash-o\"></i></a>";
                        return strDelete;
                    }
                }
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
            drawCallback: function (setting) {
                set_page_jump($(this));
            }
        }
    );//end datatable

    $(".checkall").click(function () {
        const check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('#search').click(function () {
        $("#runtimelog_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete_runtimelog();
    });
});

function delete_runtimelog(id) {
    swal({
            title: "确定要删除?",
            text: "删除当前日志记录！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/runtimelog-delete",
                {
                    "id": id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#runtimelog_table').DataTable().draw(false);
                    }
                });
        });
}

function batch_delete_runtimelog() {
    swal({
            title: "确定要清除日志?",
            text: "将批量删除满足当前查询条件的日志记录！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/runtimelog-batch-delete",
                {
                    "log_source": $('#log_source').val(),
                    "log_func": $('#log_func').val(),
                    "log_level": $('#log_level').val(),
                    "log_message": $('#log_message').val(),
                    "date_delta": $('#date_delta').val(),
                }, function (data, e) {
                    if (e === "success") {
                        $('#runtimelog_table').DataTable().draw(false);
                    }
                });
        });
}