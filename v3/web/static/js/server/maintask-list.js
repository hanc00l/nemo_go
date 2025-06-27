$(function () {
    $('#maintask_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/maintask-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "task_type": $('#task_type').val(),
                        "status": $('#task_status').val(),
                        "name": $('#task_name').val(),
                        "target": $('#task_target').val(),
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
                {
                    data: "name", title: "任务名称", width: "20%",
                    "render": function (data, type, row, meta) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + '<a href="maintask-info?id=' + row.id + '" target="' + data.task_id + '">' + data + '</a>' + '</div>';
                    }
                },
                {
                    data: "status", title: "状态", width: "5%",
                    "render": function (data, type, row, meta) {
                        if (data === "STARTED") {
                            return '<span class="badge badge-info">执行中</span>';
                        } else if (data === "SUCCESS") {
                            return '<span class="badge badge-success">成功</span>';
                        } else if (data === "FAILURE") {
                            return '<span class="badge badge-warning">失败</span>';
                        } else if (data === "CREATED") {
                            return '<span class="badge badge-secondary">已创建</span>';
                        } else if (data === "enabled") {
                            return '<span class="badge badge-info">启用</span>';
                        } else if (data === "disabled") {
                            return '<span class="badge badge-secondary">已禁用</span>';
                        } else {
                            return '<span class="badge badge-info">' + data + '</span>';
                        }
                    }
                },
                {
                    data: "progress_rate", title: "进度", width: "5%",
                    render: function (data, type, row, meta) {
                        let str_data = '';
                        if (data === "100%") str_data += '<span class="badge badge-success">100%</span>';
                        else str_data += '<span class="badge badge-info">' + data + '</span>';
                        if (isNotEmpty(row['progress'])) {
                            str_data += '</br><span title="执行中/待执行/总任务数">[' + row['progress'] + ']</span>';
                        }
                        return str_data;
                    },
                },
                {
                    data: "target", title: "目标", width: "20%", render: function (data, type, row, meta) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    }
                },
                {
                    data: "result", title: "结果", width: "15%",
                    render: function (data, type, row, meta) {
                        return '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                    },
                },
                {data: "create_time", title: "创建时间", width: "8%"},
                {data: "update_time", title: "更新时间", width: "8%"},
                {data: "runtime", title: "执行时长", width: "5%"},
                {
                    title: "操作",
                    width: "8%",
                    "render": function (data, type, row, meta) {
                        let str_data = "";
                        if (row["cron"] === false) {
                            str_data += '<a href="asset-list?task_id=' + row.task_id + '" target="' + row.task_id + '">查看</a>'
                        }
                        if (row["cron"] === true) {
                            str_data += '&nbsp;<a onclick=change_cron_status("' + row.id + '") href="#"><i class="fa fa-cog" title="更改状态"></i></a>'
                        } else {
                            str_data += '&nbsp;<a href="maintask-tree?maintaskId=' + row.task_id + '" target="' + row.task_id + '"><i class="fa fa-random" aria-hidden="true" title="查看流程图"></i></a>'
                            str_data += '&nbsp;<a onclick=redo_maintask("' + row.id + '") href="#"><i class="fa fa-reply" title="重新执行"></i></a>'
                        }
                        str_data += '&nbsp;<a onclick=delete_maintask("' + row.id + '") href="#"><i class="fa fa-trash" title="删除任务及相关资产"></i></a>'

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
        $("#maintask_table").DataTable().draw(true);
    });
    $('#maintask_add').click(function () {
        window.location.href = "maintask-add";
    })
    $('#config_profile').click(function () {
        window.location.href = "profile-list";
    })
    $("#batch_delete").click(function () {
        batch_delete('#maintask_table', '/maintask-delete');
    });
    //定时刷新页面
    setInterval(function () {
        if ($('#checkbox_auto-refresh').is(":checked")) {
            $("#maintask_table").DataTable().draw(true);
        }
    }, 15 * 1000);
});

function delete_maintask(id) {
    delete_by_id('#maintask_table', '/maintask-delete', id);
}

function change_cron_status(id) {
    swal({
            title: "确定要更改状态?",
            text: "该操作会启用或禁用该任务的定时执行, 是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认更改",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: '/maintask-change-cron-status',
                data: {id: id},
                success: function (data) {
                    show_response_message("更改定时执行状态", data, function () {
                        $("#maintask_table").DataTable().draw(false);
                    })
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
}


function redo_maintask(id) {
    swal({
            title: "确定要重新执行?",
            text: "该操作会使用当前任务的配置重新生成任务并执行, 是否继续?",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认重新执行",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.ajax({
                type: 'post',
                url: '/maintask-redo',
                data: {id: id},
                success: function (data) {
                    show_response_message("重新执行", data, function () {
                        $("#maintask_table").DataTable().draw(false);
                    })
                },
                error: function (xhr, type, error) {
                    swal('Warning', error, 'error');
                }
            });
        });
}


