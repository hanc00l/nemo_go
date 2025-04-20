$(function () {
    $('#workspace_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/workspace-list",
                "type": "post",
                "data": function (d) {
                    init_dataTables_defaultParam(d);
                    return $.extend({}, d, {});
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
                {data: "workspace_name", title: "工作空间名称", width: "10%"},
                {data: "workspace_id", title: "工作空间ID", width: "25%"},
                {data: "notify", title: "任务通知", width: "15%"},
                {data: "sort_number", title: "排序号", width: "7%"},
                {
                    data: "status", title: "状态", width: "5%",
                    "render": function (data, type, row, meta) {
                        if (data === "disable") {
                            return '<span class="badge badge-secondary">Disable</span>';
                        } else {
                            return '<span class="badge badge-success">Enable</span>';
                        }

                    }
                },
                {data: "create_time", title: "创建时间", width: "10%"},
                {data: "update_time", title: "更新时间", width: "10%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        const strModify = '<a href=workspace-edit?id=' + row.id + '><i class="fa fa-pencil"></i><span>编辑</span></a>';
                        const strDelete = '<a onclick=delete_workspace("' + row.id + '") href="#"><i class="fa fa-trash"></i><span>删除</span></a>';
                        return strModify + "&nbsp;" + strDelete;
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

    $("#workspace_update").click(function () {
        const id = $("#hidden_id").val();
        const workspace_id = $("#workspace_id").val();
        const workspace_name = $("#workspace_name").val();
        const workspace_description = $("#workspace_description").val();
        const status = $("#status").val();
        const sort_number = $("#sort_number").val();
        if (id === "" || workspace_id === "") return;
        if (workspace_name === "") {
            swal('Warning', '工作空间名称不能为空', 'error');
            return;
        }
        if (!sort_number) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/workspace-update',
            data: {
                id: id,
                workspace_id: workspace_id,
                workspace_name: workspace_name,
                sort_number: sort_number,
                status: status,
                workspace_description: workspace_description,
            },
            success: function (data) {
                show_response_message("更新工作空间", data, function () {
                    $("#workspace_table").DataTable().draw(false);
                    $('#editworkspace').modal('hide');
                })
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
});

function delete_workspace(id) {
    delete_by_id('#workspace_table', '/workspace-delete', id);
}


function edit_workspace(id) {
    $.ajax({
        type: 'post',
        url: '/workspace-get',
        data: {id: id},
        dataType: 'json',
        success: function (data) {
            $('#workspace_name').val(data.workspace_name);
            $('#sort_number').val(data.sort_number);
            $('#status').val(data.status);
            $('#workspace_id').val(data.workspace_id);
            $('#hidden_id').val(data.id);
            $('#workspace_description').val(data.workspace_description);
        },
        error: function (xhr, status, error) {
            swal('Warning', error, 'error');
        }
    });
}
