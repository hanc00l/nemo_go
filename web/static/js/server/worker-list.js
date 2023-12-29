$(function () {
    //$('#btnsiderbar').click();
    $('#worker_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/worker-list",
                "type": "post",
                "data": {start: 0, length: 100} //显示全部记录
            },
            columns: [
                {
                    data: "id",
                    width: "5%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        const strData = '<input type="checkbox" class="checkchild" value="' + row['worker_name'] + '"/>';
                        return strData;
                    }
                },
                {data: "index", title: "序号", width: "5%"},
                {data: "worker_name", title: "Worker", width: "20%"},
                {data: "worker_topic", title: "任务模式", width: "20%"},
                {data: 'create_time', title: '启动时间', width: '12%'},
                {
                    data: 'update_time', title: '心跳时间', width: '8%',
                    render: function (data, type, row, meta) {
                        if (row["heart_color"] === "green") {
                            return '<span class="text-primary">' + data + '</span>';
                        } else if (row["heart_color"] === "yellow") {
                            return '<span class="text-warning">' + data + '</span>';
                        } else return '<span class="text-danger">' + data + '</span>';
                    }
                },
                {
                    title: "<span title='正在执行/已执行'>任务</span>", width: '6%',
                    render: function (data, type, row, meta) {
                        return row["started_number"] + " / " + row["task_number"];
                    }
                },
                {
                    title: "<span title='CPU/Mem'>负载</span>", width: '12%',
                    render: function (data, type, row, meta) {
                        let strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">';
                        strData += row["cpu_load"] + " " + row["mem_used"];
                        strData += '</div>'
                        return strData
                    }
                },
                {
                    title: "操作", width: '15%',
                    render: function (data, type, row, meta) {
                        let str = "";
                        str += '&nbsp;<button class="btn btn-sm btn-primary" type="button" onclick="edit_option(\'' + row['worker_name'] + '\',' + row['daemon_process'] + ')" ><i class="fa fa-pencil"></i>参数</button>';
                        if (row["enable_manual_reload_flag"] === true) {
                            str += '&nbsp;<button class="btn btn-sm btn-primary" type="button" onclick="reload_worker(\'' + row['worker_name'] + '\')" ><i class="fa fa-play-circle"></i>重启</button>';
                        }
                        return str
                    }
                }
            ],
            infoCallback: function (settings, start, end, max, total, pre) {
                return "共<b>" + total + "</b>条记录，当前显示" + start + "到" + end + "记录";
            },
            drawCallback: function (setting) {
                var _this = $(this);
                var tableId = _this.attr('id');
                var pageDiv = $('#' + tableId + '_paginate');
                pageDiv.append(
                    '<i class="fa fa-arrow-circle-o-right fa-lg" aria-hidden="true"></i><input id="' + tableId + '_gotoPage" type="text" style="height:20px;line-height:20px;width:40px;"/>' +
                    '<a class="paginate_button" aria-controls="' + tableId + '" tabindex="0" id="' + tableId + '_goto">Go</a>')
                $('#' + tableId + '_goto').click(function (obj) {
                    var page = $('#' + tableId + '_gotoPage').val();
                    var thisDataTable = $('#' + tableId).DataTable();
                    var pageInfo = thisDataTable.page.info();
                    if (isNaN(page)) {
                        $('#' + tableId + '_gotoPage').val('');
                        return;
                    } else {
                        var maxPage = pageInfo.pages;
                        var page = Number(page) - 1;
                        if (page < 0) {
                            page = 0;
                        } else if (page >= maxPage) {
                            page = maxPage - 1;
                        }
                        $('#' + tableId + '_gotoPage').val(page + 1);
                        thisDataTable.page(page).draw('page');
                    }
                })
            }
        }
    );//end datatable
    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('[data-toggle="tooltip"]').tooltip();
    //搜索
    $("#refresh").click(function () {
        $("#worker_table").DataTable().draw(true);
    });
    $("#buttonUpdate").click(function () {
        $.post("/worker-update",
            {
                "worker_name": $('#input_worker_name').val(),
                "concurrency": $('#select_concurrency').val(),
                "worker_performance": $('#select_worker_performance').val(),
                "worker_run_task_mode": $('#select_worker_run_task_mode').val(),
                "task_workspace_guid": $('#input_task_workspace_guid').val(),
                "default_config_file": $('#input_default_config_file').val(),
                "no_proxy": $('#checkbox_no_proxy').is(":checked"),
            }, function (data, e) {
                if (e === "success" && data['status'] == 'success') {
                    swal({
                        title: "更新成功！",
                        text: data["msg"],
                        type: "success",
                        confirmButtonText: "确定",
                        confirmButtonColor: "#41b883",
                        closeOnConfirm: true,
                        timer: 3000
                    }, function () {
                        $('#editWorkerOption').modal('hide');
                    });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });
    });
    $("#batch_reload").click(function () {
        let dataTableId = "#worker_table"
        let url = "worker-reload"
        swal({
                title: "确定要批量重启选定的Worker?",
                text: "该操作会下发重启命令到所有选定Worker，由Daemon进程重启Worker；如果Worker不是由Daemon进程启动的将无法重启！",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "确认重启",
                cancelButtonText: "取消",
                closeOnConfirm: true
            },
            function () {
                $(dataTableId).DataTable().$('input[type=checkbox]:checked').each(function (i) {
                    let id = $(this).val().split("|")[0];
                    $.ajax({
                        type: 'post',
                        async: false,
                        url: url,
                        data: {"worker_name": id},
                        success: function (data) {
                        },
                        error: function (xhr, type) {
                        }
                    });
                });
                $(dataTableId).DataTable().draw(false);
            });
    });
    //定时刷新页面
    setInterval(function () {
        if ($('#checkbox_auto-refresh').is(":checked")) {
            $("#worker_table").DataTable().draw(true);
        }
    }, 10 * 1000);
});

/**
 * 移除 dataTables默认参数，并设置分页值
 * @param param
 */
function init_dataTables_defaultParam(param) {
    for (var key in param) {
        if (key.indexOf("columns") == 0 || key.indexOf("order") == 0 || key.indexOf("search") == 0) { //以columns开头的参数删除
            delete param[key];
        }
    }
    param.pageSize = param.length;
    param.pageNum = (param.start / param.length) + 1;
}

/**
 * 重启worker
 * @param worker_name
 */
function reload_worker(worker_name) {
    swal({
            title: "确定要重启worker吗?",
            text: "将重启worker命令发送至守护进程，由守护进程结束当前正在运行的worker进程，并执行一次文件同步后，启动新的worker进程！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认重启",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/worker-reload",
                {
                    "worker_name": worker_name,
                }, function (data, e) {
                    if (e === "success" && data['status'] == 'success') {
                        $('#worker-table').DataTable().draw(false);
                    }
                })
        })
}


function edit_option(worker_name, daemon_process) {
    $('#editWorkerOption').modal('toggle');
    $('#input_worker_name').val(worker_name);
    $("#buttonUpdate").prop("disabled", !daemon_process);
    $.post("/worker-edit",
        {
            "worker_name": worker_name,
        }, function (data, e) {
            if (e === "success") {
                $('#select_concurrency').val(data['concurrency']);
                $('#select_worker_performance').val(data['worker_performance']);
                $('#select_worker_run_task_mode').val(data['worker_run_task_mode']);
                $('#input_task_workspace_guid').val(data['task_workspace_guid']);
                $('#input_default_config_file').val(data['default_config_file']);
                $('#checkbox_no_proxy').prop("checked", data['no_proxy']);
            } else {
                swal('Warning', data['msg'], 'error');
            }
        })
}

