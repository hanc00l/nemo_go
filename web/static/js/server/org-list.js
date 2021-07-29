$(function () {
    $('#btnsiderbar').click();
    $('#org_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/org-list",
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
                {data: "org_name", title: "组织名称", width: "30%"},
                {data: "status", title: "状态", width: "5%"},
                {data: "sort_order", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "15%"},
                {data: "update_time", title: "更新时间", width: "15%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        var strModify = '<a onclick="edit_org(' + row.id + ')" role="button" data-toggle="modal" href="#" title="Edit" data-target="#editOrg"><i class="fa fa-pencil"></i><span>Edit</span></a>';
                        var strDelete = '<a onclick="delete_org(' + row.id + ')" href="#"><i class="fa fa-trash"></i><span>Delete</span></a>';
                        return strModify + strDelete;
                    }
                }
            ]
        }
    );//end datatable


    $(".checkall").click(function () {
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

    $("#org_update").click(function () {
        const org_id = $("#org_id").val();
        const org_name = $("#org_name").val();
        const status = $("#status").val();
        const sort_order = $("#sort_order").val();
        if (!org_id) return;
        if (!org_name) {
            swal('Warning', '组织名称不能为空', 'error');
            return;
        }
        if (!sort_order) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.post("/org-update?id=" + org_id,
            {
                "org_name": org_name,
                "sort_order": sort_order,
                'status': status
            }, function (data, e) {
                if (e === "success") {
                    swal({
                            title: "更新组织成功",
                            text: "",
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            $("#org_table").DataTable().draw(false);
                            $('#editOrg').modal('hide');
                        });
                } else {
                    swal('Warning', "更新组织失败!", 'error');
                }
            });
    });
});

function delete_org(id) {
    swal({
            title: "确定要删除?",
            text: "该操作会删除该组织的所有IP和DOMAIN资产！",
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
                url: '/org-delete?id=' + id,
                success: function (data) {
                    $("#org_table").DataTable().draw(false);
                },
                error: function (xhr, type) {
                }
            });
        });
}


function edit_org(id) {
    $.ajax({
        type: 'post',
        url: '/org-get?id=' + id,
        dataType: 'json',
        success: function (e) {
            const data = eval(e);
            const org_name = data.org_name;
            const sort_order = data.sort_order;
            const status = data.status
            $('#org_name').val(org_name);
            $('#sort_order').val(sort_order);
            $('#status').val(status);
            $('#org_id').val(id);
        },
        error: function (xhr, type) {
        }
    });
}