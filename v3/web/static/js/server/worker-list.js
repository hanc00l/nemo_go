$(function () {
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
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {});
                }
            },
            columns: [
                {
                    data: "id",
                    width: "5%",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    render: function (data, type, row) {
                        return '<input type="checkbox" class="checkchild" value="' + row['worker_name'] + '"/>';
                    }
                },
                {data: "index", title: "序号", width: "5%"},
                {
                    data: "worker_name", title: "Worker", width: "20%",
                    render: function (data, type, row, meta) {
                        let strData = data;
                        if (row["ipv6"] === true) {
                            strData += '<span class="badge badge-success" title="支持ipv6扫描">(ipv6)</span>';
                        }
                        return strData;
                    }
                },
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
                    title: "<span title='正在执行/已执行'>任务数</span>", width: '6%',
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
                set_page_jump($(this));
            }
        }
    );//end datatable

    $(".checkall").click(function () {
        const check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });
    $('[data-toggle="tooltip"]').tooltip();
    //搜索
    $("#refresh").click(function () {
        $("#worker_table").DataTable().draw(true);
    });
    $("#buttonUpdate").click(function () {
        let worker_run_task_mode = "";
        if ($('#checkbox_task_mode_0').is(":checked")) {
            worker_run_task_mode = "0"
        } else {
            let sep = "";
            for (let i = 1; i <= 4; i++) {
                if ($('#checkbox_task_mode_' + i).is(":checked")) {
                    worker_run_task_mode += sep;
                    worker_run_task_mode += i;
                    sep = ","
                }
            }
        }
        if (worker_run_task_mode === "") {
            alert("必须指定Worker的任务模式！")
            return
        }
        $('#editWorkerOption').modal('hide');
        swal({
                title: "确定要重置、更新worker吗?",
                text: "将重置、更新worker命令发送至守护进程，由守护进程结束当前正在运行的worker进程，同时会清除原有的worker文件和thirdparty目录、重新请求下载最新版本的worker文件，并使用新的配置参数启动worker进程！",
                type: "warning",
                showCancelButton: true,
                confirmButtonColor: "#DD6B55",
                confirmButtonText: "确认更新",
                cancelButtonText: "取消",
                closeOnConfirm: true
            },
            function () {
                $.post("/worker-update",
                    {
                        "service_host": $('#input_service_host').val(),
                        "service_port": $('#input_service_port').val(),
                        "service_auth": $('#input_service_auth').val(),
                        "worker_name": $('#input_worker_name').val(),
                        "concurrency": $('#select_concurrency').val(),
                        "worker_performance": $('#select_worker_performance').val(),
                        "worker_run_task_mode": worker_run_task_mode,
                        "default_config_file": $('#input_default_config_file').val(),
                        "no_proxy": $('#checkbox_no_proxy').is(":checked"),
                        "no_redis_proxy": $('#checkbox_no_redis_proxy').is(":checked"),
                        "ipv6": $('#checkbox_ip_v6').is(":checked"),
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
            })
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
 * 重启worker
 * @param worker_name
 */
function reload_worker(worker_name) {
    swal({
            title: "确定要重启worker吗?",
            text: "将重启worker命令发送至守护进程，由守护进程结束当前正在运行的worker进程，启动新的worker进程！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认重启",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: "POST",
                url: "/worker-reload",
                data: {"worker_name": worker_name},
                success: function (data) {
                    if (data['status'] === 'success') {
                        $('#worker-table').DataTable().draw(false);
                    }
                },
                error: function (xhr, status, error) {
                    // 请求失败，处理错误
                    console.error('请求失败:', error);
                    alert('请求失败: ' + error);
                }
            });
        });
}


function edit_option(worker_name, daemon_process) {
    $('#editWorkerOption').modal('toggle');
    $('#input_worker_name').val(worker_name);
    $("#buttonUpdate").prop("disabled", !daemon_process);
    $.ajax({
        type: "POST",
        url: "/worker-edit",
        data: {
            "worker_name": worker_name
        },
        success: function (data, e) {
            if (e === "success") {
                $('#input_service_host').val(data["service_host"]);
                $('#input_service_port').val(data["service_port"]);
                $('#input_service_auth').val(data["service_auth"]);
                $('#select_concurrency').val(data['concurrency']);
                $('#select_worker_performance').val(data['worker_performance']);

                // 清空所有任务模式复选框
                for (let i = 0; i <= 4; i++) {
                    $('#checkbox_task_mode_' + i).prop("checked", false);
                }

                // 根据返回的任务模式设置复选框
                data['worker_run_task_mode'].split(",").forEach(function (item) {
                    if (item === "0") {
                        $('#checkbox_task_mode_0').prop("checked", true);
                    } else {
                        $('#checkbox_task_mode_' + item).prop("checked", true);
                    }
                });

                $('#input_default_config_file').val(data['default_config_file']);
                $('#checkbox_no_proxy').prop("checked", data['no_proxy']);
                $('#checkbox_no_redis_proxy').prop("checked", data['no_redis_proxy']);
                $('#checkbox_ip_v6').prop("checked", data['ipv6']);
            } else {
                // 如果请求失败，显示警告
                swal('Warning', data['msg'], 'error');
            }
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

