$(function () {
    $('#btnsiderbar').click();
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
                {data: "workspace_name", title: "工作空间名称", width: "15%"},
                {data: "workspace_guid", title: "GUID", width: "25%"},
                {data: "state", title: "状态", width: "5%"},
                {data: "sort_order", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "15%"},
                {data: "update_time", title: "更新时间", width: "15%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        var strModify = '<a onclick="edit_workspace(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Edit" data-target="#editworkspace"><i class="fa fa-pencil"></i><span>Edit</span></a>';
                        var strDelete = '<a onclick="delete_workspace(' + row.id + ')" href="#"><i class="fa fa-trash"></i><span>Delete</span></a>';
                        return strModify + "&nbsp;" + strDelete;
                    }
                }
            ]
        }
    );//end datatable


    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

    $("#workspace_update").click(function () {
        const workspace_id = $("#workspace_id").val();
        const workspace_name = $("#workspace_name").val();
        const workspace_description = $("#workspace_description").val();
        const state = $("#state").val();
        const sort_order = $("#sort_order").val();
        if (!workspace_id) return;
        if (!workspace_name) {
            swal('Warning', '工作空间名称不能为空', 'error');
            return;
        }
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.post("/workspace-update?id=" + workspace_id,
            {
                "workspace_name": workspace_name,
                "sort_order": sort_order,
                'state': state,
                'workspace_description': workspace_description,
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "更新工作空间成功",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $("#workspace_table").DataTable().draw(false);
                            $('#editworkspace').modal('hide');
                        });
                } else {
                    swal('Warning', "更新工作空间失败!", 'error');
                }
            });
    });
});

function delete_workspace(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除该工作空间的所有IP和DOMAIN等资产！",
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
                url: '/workspace-delete?id=' + id,
                success: function (data) {
                    $("#workspace_table").DataTable().draw(false);
                },
                error: function (xhr, type) {
                }
            });
        });
}


function edit_workspace(id) {
    $.ajax({
        type: 'post',
        url: '/workspace-get?id=' + id,
        dataType: 'json',
        success: function (e) {
            const data = eval(e);
            $('#workspace_name').val(data.workspace_name);
            $('#sort_order').val(data.sort_order);
            $('#state').val(data.state);
            $('#workspace_id').val(id);
            $('#workspace_description').val(data.workspace_description);
        },
        error: function (xhr, type) {
        }
    });
}