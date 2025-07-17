$(function () {
    const urlParams = new URLSearchParams(window.location.search);
    const id = urlParams.get('id');
    if (!isNotEmpty(id)) {
        alert("任务ID未提供");
        return;
    }
    load_maintask_data(id);
    $('#list_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/maintask-executor-tasks",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "id": id,
                        "name": $('#executor_name').val(),
                        "status": $('#select_task_status').val(),
                        "worker": $('#worker_name').val(),
                    });
                }
            },
            columns: [
                {
                    data: "id",
                    className: "dt-body-center",
                    title: '<input  type="checkbox" class="checkall" />',
                    "render": function (data, type, row) {
                        return '<input type="checkbox" class="checkchild" value="' + data + '"/>';
                    }
                },
                {data: "index", title: "序号", width: "5%"},
                {data: "executor", title: "任务名称", width: "8%"},
                {
                    data: "status", title: "状态", width: "5%",
                    "render": function (data, type, row) {
                        return get_status_color(data);
                    }
                },
                {
                    data: "target", title: "目标", width: "20%",
                    "render": function (data, type, row) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    }
                },
                {
                    data: "result", title: "结果", width: "15%",
                    "render": function (data, type, row) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    }
                },
                {data: "start_time", title: "开始时间", width: "8%"},
                {data: "end_time", title: "结束时间", width: "8%"},
                {data: "runtime", title: "时长", width: "5%"},
                {
                    data: "worker", title: "Worker", width: "10%",
                    "render": function (data, type, row) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    }
                },
                {data: "create_time", title: "创建时间", width: "8%"},
                {
                    title: "操作", width: "8%",
                    "render": function (data, type, row, meta) {
                        let str_data = "";
                        str_data += '<a class="btn btn-sm btn-danger"  href="javascript:delete_executor_ask(\'' + row.id + '\')" role="button" title="删除"> <i class="fa fa-trash-o"></i> </a>';
                        str_data += '<a class="btn btn-sm btn-info"  href="javascript:redo_executor_ask(\'' + row.id + '\')" role="button" title="重新执行"> <i class="fa fa-reply"></i> </a>';
                        return str_data;
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
        $("#list_table").DataTable().draw(true);
    });
})

function fill_form_with_data(data) {
    $('#task_id').text(data.task_id);
    $("#task_name").text(data.name);
    $("#task_status").text(data.status);
    $("#task_profile_name").text(data.profile_name);
    $("#task_workspace_name").text(data.workspace);
    $("#task_org_name").text(data.org_name);
    $("#task_create_time").text(data.create_time);
    $("#task_start_time").text(data.start_time);
    $("#task_end_time").text(data.end_time);
    $("#task_run_time").text(data.runtime);
    $("#task_executors_num").text(data.progress);
    $("#task_progress").text(data.progress_rate);
    $("#task_result").text(data.result);
    $("#task_proxy").text(data.proxy);
    $("#task_cron").text(data.cron_task);
    $("#task_split").text(data.target_split);
    $("#task_description").text(data.description);
    $("#task_target").text(data.target);
    $("#task_exclude_target").text(data.exclude_target);
    $("#task_args").text(data.args);
    if (data.is_cron_task === true) {
        $('#task_executor_card').hide();
    }
    if (data.report === "success") {
        $('#task_report').html('<a href="maintask-report?task_id=' + data.task_id + '" target="_blank"><i class="fa fa-file-text-o" aria-hidden="true"></i> 查看报告(' + data.report_llmapi + ')</a>');
    } else {
        $('#task_report').text(data.report_llmapi);
    }
}

function load_maintask_data(id) {
    $.ajax({
        type: 'POST',
        url: '/maintask-info', // Beego处理URL
        data: {id: id},
        dataType: 'json',
        success: function (response) {
            fill_form_with_data(response);
        },
        error: function (xhr, status, error) {
            // 请求失败，处理错误
            console.error('请求失败:', error);
            alert('请求失败: ' + error);
        }
    });
}

function delete_executor_ask(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除该任务, 是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: '/maintask-executor-task-delete',
                data: {id: id},
                success: function (data) {
                    $("#list_table").DataTable().draw(true);
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
}

function redo_executor_ask(id, maintaskId) {
    swal({
            title: "确定要重新执行?",
            text: "该操作会重新执行该任务, 只有在主任务还未结束的时候才可以重新执行该任务。是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: '/maintask-executor-task-redo',
                data: {id: id},
                success: function (data) {
                    $("#list_table").DataTable().draw(true);
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
}

function get_status_color(status) {
    switch (status) {
        case "FAILURE":
            return '<span class="badge bg-danger" style="font-size: 12px;color: white;">失败</span>';
        case "STARTED":
            return '<span class="badge bg-info" style="font-size: 12px;color: white;">执行中</span>';
        case "SUCCESS":
            return '<span class="badge bg-success" style="font-size: 12px;color: white;">成功</span>';
        case "CREATED":
            return '<span class="badge bg-secondary" style="font-size: 12px;color: white;">已创建</span>';
        default:
            return '<span class="badge bg-secondary" style="font-size: 12px; color: white;">status</span>';
    }
}