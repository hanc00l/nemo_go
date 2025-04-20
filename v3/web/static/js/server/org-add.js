$(function () {
    $("#org_add").click(function () {
        const org_name = $("#org_name").val();
        const status = $("#status").val();
        const sort_number = $("#sort_number").val();
        const description = $("#description").val();
        if (!isNotEmpty(org_name)) {
            swal('Warning', '组织名称不能为空', 'error');
            return;
        }
        if (!isNotEmpty(sort_number)) {
            swal('Warning', '排序号不能为空', 'error');
            return;
        }
        if (!isNotEmpty(status)) {
            swal('Warning', '请指定状态', 'error');
            return;
        }
        $.ajax({
            type: 'post',
            url: '/org-add',
            data: {
                org_name: org_name,
                sort_number: sort_number,
                status: status,
                description: description,
            },
            success: function (data) {
                show_response_message("新增组织", data, function () {
                    location.href = "/org-list"
                })
            },
            error: function (xhr, type, error) {
                swal('Warning', error, 'error');
            }
        });
    });
});