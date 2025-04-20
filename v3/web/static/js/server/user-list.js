$(function () {
    //$('#btnsiderbar').click();
    $('#user_table').DataTable(
        {
            "paging": true,
            "serverSide": true,
            "autowidth": false,
            "sort": false,
            "pagingType": "full_numbers",//分页样式
            'iDisplayLength': 50,
            "dom": '<i><t><"bottom"lp>',
            "ajax": {
                "url": "/user-list",
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
                {
                    data: "user_name",
                    title: "用户名称", width: "15%",
                    "render": function (data, type, row, meta) {
                        return '<a onclick=edit_user("' + row.id + '") role="button" data-toggle="modal" href="#" title="Edit" data-target="#edituser"><span> ' + data + '</span></a>';
                    }
                },
                {data: "user_role", title: "用户类型", width: "8%"},
                {data: "user_description", title: "用户描述", width: "15%"},
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
                {data: "create_time", title: "创建时间", width: "12%"},
                {data: "update_time", title: "更新时间", width: "12%"},
                {
                    title: "操作",
                    "render": function (data, type, row, meta) {
                        const strDelete = '<a onclick=delete_user("' + row.id + '") href="#"><i class="fa fa-trash"></i><span>删除</span></a>';
                        const strResetPassword = '<a onclick=reset_user_password("' + row.id + '") role="button" data-toggle="modal" href="#" title="Reset" data-target="#resetpassword"><i class="fa fa-unlock-alt"></i><span>重置密码</span></a>';
                        const strUserWorkspace = '<a onclick=set_user_workspace("' + row.id + '") role="button" data-toggle="modal" href="#" title="Workspace" data-target="#setuserworkspace"><i class="fa fa-cube"></i><span>工作空间</span></a>';
                        return strUserWorkspace + "&nbsp;" + strResetPassword + "&nbsp;" + strDelete;
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
        var check = $(this).prop("checked");
        $(".checkchild").prop("checked", check);
    });

    $("#user_update").click(function () {
        const user_id = $("#hidden_id").val();
        const user_name = $("#user_name").val();
        const user_description = $("#user_description").val();
        const status = $("#status").val();
        const sort_number = $("#sort_number").val();
        const user_role = $("#user_role").val();
        if (!user_id) return;
        if (!user_name) {
            swal('Warning', '用户名称不能为空', 'error');
            return;
        }
        if (!sort_number) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/user-update',
            data: {
                id: user_id,
                user_name: user_name,
                sort_number: sort_number,
                status: status,
                user_description: user_description,
                user_role: user_role
            },
            success: function (data) {
                show_response_message("更新用户", data, function () {
                    $("#user_table").DataTable().draw(false);
                    $('#edituser').modal('hide');
                });
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });

    $("#reset_user_password").click(function () {
        const user_id = $("#reset_user_id").val();
        const user_password1 = $("#reset_user_password1").val();
        const user_password2 = $("#reset_user_password2").val();
        if (!user_id || !user_password1 || !user_password2) return;
        if (user_password1 !== user_password2) {
            swal('Warning', '两次密码不一致！', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/user-reset-password',
            data: {
                id: user_id,
                user_password: user_password1,
            },
            success: function (data) {
                show_response_message("重置密码", data, function () {
                    $('#resetpassword').modal('hide');
                });
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });

    $("#set_user_workspace").click(function () {
        const user_id = $("#userworkspace_user_id").val();
        if (!user_id) return;

        let selected_workspace_id = "";
        const checkSelected = $('input:checkbox[name="cb_userworkspace"]');
        for (let i = 0; i < checkSelected.length; i++) {
            if (checkSelected[i].checked) {
                if (checkSelected[i].value !== '') {
                    if (selected_workspace_id === "") selected_workspace_id = checkSelected[i].value;
                    else selected_workspace_id += ',' + checkSelected[i].value;
                }
            }
        }
        $.ajax({
            type: 'post',
            url: '/user-workspace-update',
            data: {
                user_id: user_id,
                workspace_id: selected_workspace_id
            },
            success: function (data) {
                show_response_message("更新用户工作空间", data, function () {
                    $('#setuserworkspace').modal('hide');
                });
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
});

function delete_user(id) {
    delete_by_id('#user_table','/user-delete',id);
}


function edit_user(id) {
    $.ajax({
        type: 'post',
        url: '/user-get',
        data: {id: id},
        dataType: 'json',
        success: function (data) {
            $('#user_name').val(data.user_name);
            $('#sort_number').val(data.sort_number);
            $('#status').val(data.status);
            $('#hidden_id').val(id);
            $('#user_description').val(data.user_description);
            $('#user_role').val(data.user_role);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}

function reset_user_password(id) {
    $('#reset_user_id').val(id);
}

function set_user_workspace(id) {
    $('#userworkspace_user_id').val(id);
    $('#checkbox_user_workspace_list').empty();
    $.ajax({
        type: 'post',
        url: '/user-workspace-list',
        data: {
            user_id: id,
        },
        success: function (data) {
            let strCheckBox = '';
            for (let i = 0; i < data.length; i++) {
                strCheckBox += '<div className="form-check"><input className="form-check-input" type="checkbox" value="';
                strCheckBox += data[i].workspace_id;
                strCheckBox += '" name="cb_userworkspace" id="cb_userworkspace_';
                strCheckBox += data[i].workspace_id;
                strCheckBox += '"';
                if (data[i].enable === true) strCheckBox += ' checked';
                strCheckBox += '><label className="form-check-label" for="cb_userworkspace_';
                strCheckBox += data[i].workspace_id;
                strCheckBox += '">&nbsp;';
                strCheckBox += data[i].workspace_name;
                strCheckBox += '</label></div>';
            }
            $('#checkbox_user_workspace_list').append(strCheckBox);
        },
        error: function (xhr, type, error) {
            swal('Warning', error, 'error');
        }
    });
}