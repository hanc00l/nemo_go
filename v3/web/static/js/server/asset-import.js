$(function () {
    load_select_list('/org-all-list', $('#org_id'));
    $('#btn-import').click(function () {
        import_asset();
    });

    function import_asset() {
        const org_id = $('#org_id').val();
        const bin = $('#bin').val()
        const file = $('#file')[0].files[0];
        const insert_or_update = $('#insert_and_update').prop('checked');
        const form_data = new FormData();
        form_data.append('file', file);
        form_data.append('bin', bin);
        form_data.append('org_id', org_id);
        form_data.append('insert_and_update',insert_or_update );
        $.ajax({
            url: '/asset-import',
            type: 'POST',
            data: form_data,
            contentType: false,
            processData: false,
            success: function (data) {
                show_response_message("导入资产", data, function () {
                });
            },
            error: function (xhr, status, error) {
                alert('Asset import failed');
            }
        });
    }
});