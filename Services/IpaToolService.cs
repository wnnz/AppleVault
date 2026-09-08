using System.Diagnostics;
using System.IO;
using System.Text;
using System.Text.Json;
using System.Text.RegularExpressions;
using IPAToolGUI.Models;

namespace IPAToolGUI.Services;

public class ExecutionResult
{
    public bool Success { get; set; }
    public int ExitCode { get; set; }
    public string StandardOutput { get; set; } = string.Empty;
    public string StandardError { get; set; } = string.Empty;
    public string ErrorMessage { get; set; } = string.Empty;
}

public class IpaToolService
{
    private readonly SettingsService _settingsService;

    public event Action<string>? LogReceived;

    public IpaToolService(SettingsService settingsService)
    {
        _settingsService = settingsService;
    }

    private void Log(string message)
    {
        LogReceived?.Invoke($"[{DateTime.Now:HH:mm:ss}] {message}");
    }

    public async Task<ExecutionResult> ExecuteAsync(List<string> args, CancellationToken ct = default)
    {
        var exePath = SettingsService.ResolveIpaToolPath();
        if (string.IsNullOrWhiteSpace(exePath) || !File.Exists(exePath))
        {
            var err = $"未在程序同目录下找到 ipatool.exe！请确认 ipatool.exe 存在于: \"{AppDomain.CurrentDomain.BaseDirectory}\"";
            Log(err);
            return new ExecutionResult { Success = false, ErrorMessage = err };
        }

        // Add global flags if not already provided
        var commandArgs = new List<string>(args);

        if (!commandArgs.Contains("--non-interactive"))
        {
            commandArgs.Add("--non-interactive");
        }

        var passphrase = _settingsService.Settings.KeychainPassphrase;
        if (!string.IsNullOrEmpty(passphrase) && !commandArgs.Any(a => a.StartsWith("--keychain-passphrase")))
        {
            commandArgs.Add("--keychain-passphrase");
            commandArgs.Add(passphrase);
        }

        var psi = new ProcessStartInfo
        {
            FileName = exePath,
            UseShellExecute = false,
            RedirectStandardOutput = true,
            RedirectStandardError = true,
            CreateNoWindow = true,
            StandardOutputEncoding = Encoding.UTF8,
            StandardErrorEncoding = Encoding.UTF8
        };

        // Inject Proxy Environment Variables
        if (_settingsService.Settings.EnableProxy && !string.IsNullOrWhiteSpace(_settingsService.Settings.ProxyUrl))
        {
            var proxy = _settingsService.Settings.ProxyUrl.Trim();
            psi.EnvironmentVariables["HTTP_PROXY"] = proxy;
            psi.EnvironmentVariables["HTTPS_PROXY"] = proxy;
            psi.EnvironmentVariables["ALL_PROXY"] = proxy;
            psi.EnvironmentVariables["http_proxy"] = proxy;
            psi.EnvironmentVariables["https_proxy"] = proxy;
            psi.EnvironmentVariables["all_proxy"] = proxy;
        }
        else
        {
            psi.EnvironmentVariables.Remove("HTTP_PROXY");
            psi.EnvironmentVariables.Remove("HTTPS_PROXY");
            psi.EnvironmentVariables.Remove("ALL_PROXY");
            psi.EnvironmentVariables.Remove("http_proxy");
            psi.EnvironmentVariables.Remove("https_proxy");
            psi.EnvironmentVariables.Remove("all_proxy");
        }

        foreach (var arg in commandArgs)
        {
            psi.ArgumentList.Add(arg);
        }

        // Hide passphrase from user log
        var displayArgs = string.Join(" ", commandArgs.Select(a =>
            a.Contains(" ") ? $"\"{a}\"" : a));
        if (!string.IsNullOrEmpty(passphrase))
        {
            displayArgs = displayArgs.Replace(passphrase, "******");
        }

        if (_settingsService.Settings.EnableProxy && !string.IsNullOrWhiteSpace(_settingsService.Settings.ProxyUrl))
        {
            Log($"[网络代理] 已启用: {_settingsService.Settings.ProxyUrl}");
        }

        Log($"> ipatool {displayArgs}");

        var stdoutBuilder = new StringBuilder();
        var stderrBuilder = new StringBuilder();

        using var process = new Process { StartInfo = psi };

        process.OutputDataReceived += (_, e) =>
        {
            if (e.Data != null)
            {
                stdoutBuilder.AppendLine(e.Data);
                Log(e.Data);
            }
        };

        process.ErrorDataReceived += (_, e) =>
        {
            if (e.Data != null)
            {
                stderrBuilder.AppendLine(e.Data);
                Log($"[ERR] {e.Data}");
            }
        };

        try
        {
            process.Start();
            process.BeginOutputReadLine();
            process.BeginErrorReadLine();

            await process.WaitForExitAsync(ct);

            var stdout = stdoutBuilder.ToString().Trim();
            var stderr = stderrBuilder.ToString().Trim();
            var success = process.ExitCode == 0;

            string errMsg = string.Empty;
            if (!success)
            {
                errMsg = ExtractErrorMessage(stderr, stdout);
            }

            return new ExecutionResult
            {
                Success = success,
                ExitCode = process.ExitCode,
                StandardOutput = stdout,
                StandardError = stderr,
                ErrorMessage = errMsg
            };
        }
        catch (OperationCanceledException)
        {
            try { process.Kill(true); } catch { }
            Log("操作已取消。");
            return new ExecutionResult { Success = false, ErrorMessage = "操作已被用户取消。" };
        }
        catch (Exception ex)
        {
            Log($"执行异常: {ex.Message}");
            return new ExecutionResult { Success = false, ErrorMessage = ex.Message };
        }
    }

    private static string ExtractErrorMessage(string stderr, string stdout)
    {
        // Try extracting error="..." from stderr / stdout
        var match = Regex.Match(stderr + "\n" + stdout, @"error=""([^""]+)""");
        if (match.Success)
        {
            return match.Groups[1].Value;
        }

        if (!string.IsNullOrWhiteSpace(stderr))
        {
            return stderr;
        }

        if (!string.IsNullOrWhiteSpace(stdout))
        {
            return stdout;
        }

        return "未知错误，进程返回非零退出码。";
    }

    private static T? TryParseJson<T>(string output) where T : class
    {
        // Sometimes zerolog logs lines alongside json, so find the JSON object line
        var lines = output.Split(new[] { '\r', '\n' }, StringSplitOptions.RemoveEmptyEntries);
        for (int i = lines.Length - 1; i >= 0; i--)
        {
            var line = lines[i].Trim();
            if (line.StartsWith("{") && line.EndsWith("}"))
            {
                try
                {
                    var res = JsonSerializer.Deserialize<T>(line);
                    if (res != null) return res;
                }
                catch
                {
                    // Continue searching
                }
            }
        }

        // Try the whole string
        try
        {
            return JsonSerializer.Deserialize<T>(output);
        }
        catch
        {
            return null;
        }
    }

    // --- Command Wrappers ---

    public async Task<(bool success, AccountInfo? account, string error)> GetAccountInfoAsync(CancellationToken ct = default)
    {
        var result = await ExecuteAsync(new List<string> { "auth", "info", "--format", "json" }, ct);
        if (!result.Success)
        {
            return (false, null, result.ErrorMessage);
        }

        var account = TryParseJson<AccountInfo>(result.StandardOutput);
        return (account != null && account.Success, account, result.ErrorMessage);
    }

    public class LoginResult
    {
        public bool Success { get; set; }
        public bool Requires2FA { get; set; }
        public AccountInfo? Account { get; set; }
        public string ErrorMessage { get; set; } = string.Empty;
    }

    public async Task<LoginResult> LoginAsync(string email, string password, string authCode = "", CancellationToken ct = default)
    {
        var args = new List<string>
        {
            "auth", "login",
            "-e", email,
            "-p", password,
            "--format", "json"
        };

        if (!string.IsNullOrWhiteSpace(authCode))
        {
            args.Add("--auth-code");
            args.Add(authCode.Trim());
        }

        var result = await ExecuteAsync(args, ct);

        // Check if 2FA is required
        var allOutput = result.StandardOutput + "\n" + result.StandardError + "\n" + result.ErrorMessage;
        bool requires2FA = allOutput.Contains("2FA", StringComparison.OrdinalIgnoreCase) ||
                           allOutput.Contains("auth code is required", StringComparison.OrdinalIgnoreCase) ||
                           allOutput.Contains("auth-code", StringComparison.OrdinalIgnoreCase);

        var account = TryParseJson<AccountInfo>(result.StandardOutput);
        bool success = result.Success && account != null && account.Success;

        if (success)
        {
            return new LoginResult { Success = true, Account = account };
        }

        if (requires2FA)
        {
            return new LoginResult { Success = false, Requires2FA = true, ErrorMessage = "需要双重认证验证码" };
        }

        return new LoginResult { Success = false, Requires2FA = false, ErrorMessage = result.ErrorMessage };
    }

    public async Task<(bool success, string error)> RevokeAsync(CancellationToken ct = default)
    {
        var result = await ExecuteAsync(new List<string> { "auth", "revoke", "--format", "json" }, ct);
        return (result.Success, result.ErrorMessage);
    }

    public async Task<(bool success, List<AppItem> apps, string error)> SearchAsync(string term, int limit = 10, string platform = "", CancellationToken ct = default)
    {
        var args = new List<string>
        {
            "search", term,
            "-l", limit.ToString(),
            "--format", "json"
        };

        if (!string.IsNullOrWhiteSpace(platform))
        {
            args.Add("--platform");
            args.Add(platform);
        }

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, new List<AppItem>(), result.ErrorMessage);
        }

        var searchResult = TryParseJson<SearchResult>(result.StandardOutput);
        return (true, searchResult?.Apps ?? new List<AppItem>(), string.Empty);
    }

    public async Task<(bool success, List<string> versionIds, string error)> ListVersionsAsync(string bundleId, long appId = 0, CancellationToken ct = default)
    {
        var args = new List<string> { "list-versions", "--format", "json" };
        if (!string.IsNullOrWhiteSpace(bundleId))
        {
            args.Add("-b");
            args.Add(bundleId);
        }
        else if (appId > 0)
        {
            args.Add("-i");
            args.Add(appId.ToString());
        }
        else
        {
            return (false, new List<string>(), "必须提供 Bundle ID 或 App ID");
        }

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, new List<string>(), result.ErrorMessage);
        }

        var versionsResult = TryParseJson<VersionsResult>(result.StandardOutput);
        return (true, versionsResult?.ExternalVersionIdentifiers ?? new List<string>(), string.Empty);
    }

    public async Task<(bool success, VersionMetadataResult? meta, string error)> GetVersionMetadataAsync(string bundleId, string versionId, long appId = 0, CancellationToken ct = default)
    {
        var args = new List<string>
        {
            "get-version-metadata",
            "--external-version-id", versionId,
            "--format", "json"
        };

        if (!string.IsNullOrWhiteSpace(bundleId))
        {
            args.Add("-b");
            args.Add(bundleId);
        }
        else if (appId > 0)
        {
            args.Add("-i");
            args.Add(appId.ToString());
        }

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, null, result.ErrorMessage);
        }

        var meta = TryParseJson<VersionMetadataResult>(result.StandardOutput);
        return (meta != null && meta.Success, meta, result.ErrorMessage);
    }

    public async Task<(bool success, DownloadResult? download, string error)> DownloadAsync(
        string bundleId,
        long appId = 0,
        string externalVersionId = "",
        string outputPath = "",
        string platform = "",
        bool purchase = false,
        CancellationToken ct = default)
    {
        var args = new List<string> { "download", "--format", "json" };

        if (!string.IsNullOrWhiteSpace(bundleId))
        {
            args.Add("-b");
            args.Add(bundleId);
        }
        else if (appId > 0)
        {
            args.Add("-i");
            args.Add(appId.ToString());
        }

        if (!string.IsNullOrWhiteSpace(externalVersionId))
        {
            args.Add("--external-version-id");
            args.Add(externalVersionId);
        }

        if (!string.IsNullOrWhiteSpace(outputPath))
        {
            args.Add("-o");
            args.Add(outputPath);
        }

        if (!string.IsNullOrWhiteSpace(platform))
        {
            args.Add("--platform");
            args.Add(platform);
        }

        if (purchase)
        {
            args.Add("--purchase");
        }

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, null, result.ErrorMessage);
        }

        var dl = TryParseJson<DownloadResult>(result.StandardOutput);
        return (dl != null && dl.Success, dl, result.ErrorMessage);
    }

    public async Task<(bool success, PurchaseResult? purchase, string error)> PurchaseAsync(string bundleId, CancellationToken ct = default)
    {
        var args = new List<string>
        {
            "purchase",
            "-b", bundleId,
            "--format", "json"
        };

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, null, result.ErrorMessage);
        }

        var p = TryParseJson<PurchaseResult>(result.StandardOutput);
        return (p != null && p.Success, p, result.ErrorMessage);
    }

    public async Task<(bool success, List<AppItem> apps, int totalCount, string error)> ListPurchasesAsync(int page = 1, int limit = 20, CancellationToken ct = default)
    {
        var args = new List<string>
        {
            "list-purchases",
            "-p", page.ToString(),
            "-l", limit.ToString(),
            "--format", "json"
        };

        var result = await ExecuteAsync(args, ct);
        if (!result.Success)
        {
            return (false, new List<AppItem>(), 0, result.ErrorMessage);
        }

        var res = TryParseJson<PurchasedResult>(result.StandardOutput);
        return (true, res?.Apps ?? new List<AppItem>(), res?.TotalCount ?? 0, string.Empty);
    }
}
