package cc.juho.wgcli;

import static android.content.ContentValues.TAG;

import android.content.Intent;
import android.net.Uri;
import android.os.Bundle;
import android.text.Editable;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import android.widget.Toast;

import androidx.activity.EdgeToEdge;
import androidx.activity.result.ActivityResultLauncher;
import androidx.activity.result.contract.ActivityResultContracts;
import androidx.appcompat.app.AppCompatActivity;
import androidx.core.graphics.Insets;
import androidx.core.view.ViewCompat;
import androidx.core.view.WindowInsetsCompat;

import org.jetbrains.annotations.NotNull;

import java.io.BufferedInputStream;
import java.io.BufferedOutputStream;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;

import gomobile.CoreLogger;
import gomobile.FileHandler;
import gomobile.Gomobile;

public class MainActivity extends AppCompatActivity implements View.OnClickListener, CoreLogger, FileHandler {
    private EditText region, hour;
    private Button deploy;
    private TextView output;

    private ActivityResultLauncher<Intent> filePickerLauncher, saveFileLauncher;
    private @NotNull String clientConfSrc = "";

    // 请求码，用于标识文件保存的 Intent 回调
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        EdgeToEdge.enable(this);
        setContentView(R.layout.activity_main);
        ViewCompat.setOnApplyWindowInsetsListener(findViewById(R.id.main), (v, insets) -> {
            Insets systemBars = insets.getInsets(WindowInsetsCompat.Type.systemBars());
            v.setPadding(systemBars.left, systemBars.top, systemBars.right, systemBars.bottom);
            return insets;
        });

        //ui
        region = findViewById(R.id.region);
        hour = findViewById(R.id.hour);
        deploy = findViewById(R.id.deploy);
        output = findViewById(R.id.output);

        deploy.setOnClickListener(this);
        filePickerLauncher = registerForActivityResult(
                new ActivityResultContracts.StartActivityForResult(),
                result -> {
                    if (result.getResultCode() == RESULT_OK && result.getData() != null) {
                        // 获取选择文件的Uri
                        Uri fileUri = result.getData().getData();
                        if (fileUri != null) {
                            handleSelectedFile(fileUri);
                        }
                    } else {
                        Log.d(TAG, "未选择文件或选择取消");
                    }
                }
        );
        saveFileLauncher = registerForActivityResult(new ActivityResultContracts.StartActivityForResult(), result -> {
            if (result.getResultCode() == RESULT_OK && result.getData() != null) {
                // 获取选择文件的Uri
                Uri fileUri = result.getData().getData();
                if (fileUri == null) return;

                //save file
                OutputStream outputStream = null;
                InputStream inputStream = null;
                try {
                    outputStream = getContentResolver().openOutputStream(fileUri);
                    File fi = new File(clientConfSrc);
                    inputStream = new BufferedInputStream(new FileInputStream(fi));
                    // 5. 逐字节读取并写入文件
                    byte[] buffer = new byte[1024]; // 1KB缓冲区，提升读写效率
                    int bytesRead;
                    while ((bytesRead = inputStream.read(buffer)) != -1) {
                        assert outputStream != null;
                        outputStream.write(buffer, 0, bytesRead);
                    }
                } catch (Exception e) {
                    try {
                        if (outputStream != null)
                            outputStream.close();
                        if (inputStream != null)
                            inputStream.close();
                    } catch (IOException ex) {
                        Log.d(TAG, "onConfFileSaved: " + ex.getLocalizedMessage());
                    }
                    makeToast("save client.conf failed:" + e.getLocalizedMessage());
                }
                makeToast(getString(R.string.client_conf_saved, fileUri.getPath()));
            } else {
                Log.d(TAG, "未选择文件或选择取消");
            }
        });

        // go mobile init
        Gomobile.setLogger(this);
        Gomobile.setConfigDir(getDataDir().getAbsolutePath());
        Gomobile.setCacheDir(getCacheDir().getAbsolutePath());
        refreshState();

    }


    private void refreshState() {
        deploy.setEnabled(true);
        if (Gomobile.hasAccessKey()) {
            deploy.setText(getString(R.string.deploy));
        } else {
            deploy.setText(getString(R.string.import_access_key));
        }
    }

    @Override
    public void onClick(View v) {
        if (!Gomobile.hasAccessKey()) {
            refreshState();
            //open file picker
            openFilePicker();
            return;
        }
        deploy.setEnabled(false);
        // deploy
        new Thread(() -> {
            try {
                String r = getTextOf(region);
                Gomobile.setRegionName(r);
                Gomobile.deploy(getHour(), getCacheDir().getAbsolutePath(), this);
            } catch (Exception e) {
                makeToast(e.getLocalizedMessage());
            }
            runOnUiThread(() -> {
                deploy.setEnabled(true);
            });
        }).start();
    }

    private long getHour() {
        Editable text = hour.getText();
        if (text == null)
            return 0;
        try {
            return Long.parseLong(text.toString());
        } catch (Exception e) {
            return 0;
        }
    }

    private @NotNull String getTextOf(EditText editText) {
        Editable text = editText.getText();
        if (text == null) {
            return "";
        }
        return text.toString();
    }

    @Override
    public long write(byte[] bytes) throws Exception {
        runOnUiThread(() -> {
            CharSequence text = output.getText();
            if (text == null) {
                text = "";
            }
            text += "\n" + new String(bytes, StandardCharsets.UTF_8);
            output.setText(text);
        });
        return 0;
    }

    private void makeToast(String s) {
        if (s == null) {
            return;
        }
        runOnUiThread(() -> Toast.makeText(this, s, Toast.LENGTH_LONG).show());
    }

    private void openFilePicker() {
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        intent.setType("*/*");

        filePickerLauncher.launch(intent);
    }

    private void handleSelectedFile(Uri fileUri) {
        new Thread(() -> {

            InputStream inputStream = null;
            OutputStream outputStream = null;

            try {
                File targetFile = new File(getCacheDir(), "AccessKey.csv");
                // 3. 打开Uri对应的输入流
                inputStream = new BufferedInputStream(getContentResolver().openInputStream(fileUri));
                // 4. 打开目标文件的输出流
                outputStream = new BufferedOutputStream(new FileOutputStream(targetFile));

                // 5. 逐字节读取并写入文件
                byte[] buffer = new byte[1024]; // 1KB缓冲区，提升读写效率
                int bytesRead;
                while ((bytesRead = inputStream.read(buffer)) != -1) {
                    outputStream.write(buffer, 0, bytesRead);
                }

                // 6. 刷新输出流，确保数据全部写入
                outputStream.flush();

                // 校验文件是否保存成功
                if (targetFile.exists() && targetFile.length() > 0) {
                    String savePath = targetFile.getAbsolutePath();

                    // 可选：读取保存后的文件内容（验证）
                    Gomobile.importAccessKeyFile(targetFile.getAbsolutePath());
                    runOnUiThread(this::refreshState);
                    Log.d(TAG, "CSV文件保存成功，路径：" + savePath);
                    makeToast("文件已保存到：" + savePath);
                } else {
                    Log.e(TAG, "CSV文件保存失败，文件为空或不存在");
                    makeToast("文件保存失败");
                }

            } catch (Exception e) {
                Log.e(TAG, "保存CSV文件时发生异常：" + e.getMessage(), e);
                makeToast("保存失败：" + e.getMessage());
            } finally {
                // 7. 关闭流，释放资源
                try {
                    if (inputStream != null) inputStream.close();
                    if (outputStream != null) outputStream.close();
                } catch (IOException e) {
                    Log.e(TAG, "关闭流时发生异常：" + e.getMessage());
                }
            }
        }).start();
    }

    @Override
    public void onConfFileSaved(final String src) {
        runOnUiThread(() -> {
            clientConfSrc = src;
            Intent intent = new Intent(Intent.ACTION_CREATE_DOCUMENT);
            intent.setType("application/octet-stream");
            intent.putExtra(Intent.EXTRA_TITLE, "client.conf");
            saveFileLauncher.launch(intent);
        });
    }
}