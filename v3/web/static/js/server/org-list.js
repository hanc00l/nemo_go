$(function () {
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
                {data: "org_name", title: "组织名称", width: "25%"},
                {data: "id", title: "ID", width: "8%"},
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
                {data: "sort_number", title: "排序", width: "8%"},
                {data: "create_time", title: "创建时间", width: "15%"},
                {data: "update_time", title: "更新时间", width: "15%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        const strModify = '<a onclick=edit_org("' + row.id + '") role="button" data-toggle="modal" href="#" title="Edit" data-target="#editOrg"><i class="fa fa-pencil"></i><span>Edit</span></a>';
                        const strDelete = '<a onclick=delete_org("' + row.id + '") href="#"><i class="fa fa-trash"></i><span>Delete</span></a>';
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

    $("#org_update").click(function () {
        const org_id = $("#hidden_id").val();
        const org_name = $("#org_name").val();
        const description = $("#description").val();
        const status = $("#status").val();
        const sort_number = $("#sort_number").val();
        if (!isNotEmpty(org_id)) return;
        if (!isNotEmpty(org_name)) {
            swal('Warning', '组织名称不能为空', 'error');
            return;
        }
        if (!isNotEmpty(sort_number)) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/org-update',
            data: {
                id: org_id,
                org_name: org_name,
                description: description,
                sort_number: sort_number,
                status: status
            },
            success: function (data) {
                show_response_message("更新组织", data, function () {
                    $("#org_table").DataTable().draw(false);
                    $('#editOrg').modal('hide');
                })
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
});

function delete_org(id) {
    delete_by_id('#org_table', '/org-delete', id);
}

function edit_org(id) {
    $.ajax({
        type: 'post',
        url: '/org-get',
        data: {id: id},
        dataType: 'json',
        success: function (data) {
            $('#org_name').val(data.org_name);
            $('#sort_number').val(data.sort_number);
            $('#status').val(data.status);
            $('#description').val(data.description);
            $('#hidden_id').val(data.id);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}