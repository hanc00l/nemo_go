$(function () {
    $('#btnsiderbar').click();
    $('#task_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/task-cron-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {
                        "task_status": $('#task_status').val(),
                        "task_name": $('#task_name').val(),
                        "task_args": $('#task_args').val(),
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
                        var strData = '<input type="checkbox" class="checkchild" value="' + row['id'] + '"/>';
                        return strData;
                    }
                },
                {
                    data: "index",
                    title: "序号",
                    width: "5%"
                },
                {
                    data: "task_name",
                    title: "任务名称/说明",
                    width: "12%",
                    render: function (data, type, row, meta) {
                        let strData;
                        strData = '<a href="/task-cron-info?task_id=' + row['task_id'] + '" target="_blank">' + data + '</a>';
                        if (row['comment']) {
                            strData += "<br>" + row['comment'];
                        }
                        return strData;
                    }
                },
                {
                    data: 'status', title: '状态', width: '8%',
                    "render": function (data, type, row) {
                        if (data == 'enable') {
                            return 'Enabled' + '<button class="btn btn-sm btn-danger" type="button" onclick="disable_task(\'' + row['task_id'] + '\')" >&nbsp;禁用&nbsp;</button>';
                        } else {
                            return 'Disabled' + '<button class="btn btn-sm btn-success" type="button" onclick="enable_task(\'' + row['task_id'] + '\')" >&nbsp;启用&nbsp;</button>';
                        }
                    }
                },
                {
                    data: 'kwargs', title: '参数', width: '16%',
                    "render": function (data, type, row) {
                        const strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                        return strData;
                    }
                },
                {
                    data: 'cron_rule', title: '任务定时规则', width: '10%',
                    "render": function (data, type, row) {
                        const strData = '<div style="width:100%;white-space:normal;word-wrap:break-word;word-break:break-all;">' + data + '</div>';
                        return strData;
                    }
                },
                {data: 'create_time', title: '创建时间', width: '8%'},
                {data: 'lastrun_time', title: '最近执行时间', width: '8%'},
                {
                    data: 'run_count', title: '次数', width: '6%',
                    "render": function (data, type, row, meta) {
                        let strData;
                        if (data > 0) {
                            strData = '<a href="/task-list?cron_id=' + row['task_id'] + '" target="_blank">' + data + '</a>';
                            return strData;
                        } else {
                            return "";
                        }
                    }
                },
                {data: 'nextrun_time', title: '下次执行时间', width: '8%'},
                {
                    title: "操作",
                    width: "8%",
                    "render": function (data, type, row, meta) {
                        const strRun = "<a class=\"btn btn-sm btn-primary\" href=javascript:run_task(\"" + row["task_id"] + "\") role=\"button\" title=\"立即执行一次\"><i class=\"fa fa-play\"></i></a>";
                        const strDelete = "&nbsp;<a class=\"btn btn-sm btn-danger\" href=javascript:delete_task(\"" + row["id"] + "\") role=\"button\" title=\"Delete\"><i class=\"fa fa-trash-o\"></i></a>";
                        return strRun + strDelete;
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
    $("#search").click(function () {
        $("#task_table").DataTable().draw(true);
    });
    //批量删除
    $("#batch_delete").click(function () {
        batch_delete('#task_table', '/task-cron-delete');
    });
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
 * 禁用一个任务
 * @param task_id
 */
function disable_task(task_id) {
    swal({
            title: "确定要禁用任务?",
            text: "禁用任务！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认禁用",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/task-cron-disable",
                {
                    "task_id": task_id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#task_table').DataTable().draw(false);
                    }
                });
        });
}

/**
 * 启用一个任务
 * @param task_id
 */
function enable_task(task_id) {
    swal({
            title: "确定要启用任务?",
            text: "启用任务！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认启用",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/task-cron-enable",
                {
                    "task_id": task_id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#task_table').DataTable().draw(false);
                    }
                });
        });
}

/**
 * 删除一个任务
 * @param id
 */
function delete_task(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除当前任务，请确保当前任务已完成或中止！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/task-cron-delete",
                {
                    "id": id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#task_table').DataTable().draw(false);
                    }
                });
        });
}

//批量删除
function batch_delete(dataTableId, url) {
    swal({
            title: "确定要批量删除选定的目标?",
            text: "该操作会删除所有选定目标的所有信息！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认删除",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $(dataTableId).DataTable().$('input[type=checkbox]:checked').each(function (i) {
                let id = $(this).val().split("|")[0];
                $.ajax({
                    type: 'post',
                    url: url + '?id=' + id,
                    success: function (data) {
                    },
                    error: function (xhr, type) {
                    }
                });
            });
            $(dataTableId).DataTable().draw(false);
        });
}

/**
 * 立即执行一次任务
 * @param task_id
 */
function run_task(task_id) {
    swal({
            title: "确定要立即执行一次该任务?",
            text: "立即执行任务！",
            type: "warning",
            showCancelButton: true,
            confirmButtonColor: "#DD6B55",
            confirmButtonText: "确认执行",
            cancelButtonText: "取消",
            closeOnConfirm: true
        },
        function () {
            $.post("/task-cron-run",
                {
                    "task_id": task_id,
                }, function (data, e) {
                    if (e === "success") {
                        $('#task_table').DataTable().draw(false);
                    }
                });
        });
}