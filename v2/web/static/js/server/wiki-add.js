$(function () {
    $("#document_add").click(function () {
        const document_title = $("#document_title").val();

        if (!document_title) {
            swal('Warning', '标题不能为空', 'error');
            return;
        }

        $.post("/wiki-add",
            {
                "title": document_title,
                "comment": $('#document_comment').val(),
                "pin_index": $('#checkbox_pin_to_top').prop('checked') ? 1 : 0,
            }, function (data, e) {
                if (e === "success" && data['status'] === 'success') {
                    swal({
                            title: "新建文档成功",
                            text: data['msg'],
                            type: "success",
                            confirmButtonText: "确定",
                            confirmButtonColor: "#41b883",
                            closeOnConfirm: true,
                            timer: 3000
                        },
                        function () {
                            location.href = "/wiki-list"
                        });
                } else {
                    swal('Warning', data['msg'], 'error');
                }
            });

    });
});