using System.IO;
using System.Text.Json;
using IPAToolGUI.Models;

namespace IPAToolGUI.Services;

public class SettingsService
{
    private readonly string _settingsFilePath;
    private static readonly JsonSerializerOptions JsonOptions = new() { WriteIndented = true };

    public AppSettings Settings { get; private set; } = new();

    public SettingsService()
    {
        var appDataDir = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData), "IPAToolGUI");
        if (!Directory.Exists(appDataDir))
        {
            Directory.CreateDirectory(appDataDir);
        }
        _settingsFilePath = Path.Combine(appDataDir, "settings.json");
        Load();
    }

    public static string ResolveIpaToolPath()
    {
        // 1. Same directory as current executable
        var baseDir = AppDomain.CurrentDomain.BaseDirectory;
        var inBaseDir = Path.Combine(baseDir, "ipatool.exe");
        if (File.Exists(inBaseDir)) return inBaseDir;

        var inBaseDirLong = Path.Combine(baseDir, "ipatool-2.5.0-windows-amd64.exe");
        if (File.Exists(inBaseDirLong)) return inBaseDirLong;

        // 2. Project directory or downloads fallback
        var downloadPath = @"D:\Downloads\ipatool-2.5.0-windows-amd64\bin\ipatool.exe";
        if (File.Exists(downloadPath)) return downloadPath;

        var devPath = @"D:\Dev\IPAToolGUI\ipatool.exe";
        if (File.Exists(devPath)) return devPath;

        return "ipatool.exe";
    }

    public void Load()
    {
        if (File.Exists(_settingsFilePath))
        {
            try
            {
                var json = File.ReadAllText(_settingsFilePath);
                var loaded = JsonSerializer.Deserialize<AppSettings>(json);
                if (loaded != null)
                {
                    Settings = loaded;
                }
            }
            catch
            {
                // Use default if corrupted
            }
        }

        if (string.IsNullOrWhiteSpace(Settings.DefaultDownloadDir) || !Directory.Exists(Settings.DefaultDownloadDir))
        {
            var downloads = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), "Downloads");
            if (Directory.Exists(@"D:\Downloads"))
            {
                downloads = @"D:\Downloads";
            }
            Settings.DefaultDownloadDir = downloads;
        }

        if (string.IsNullOrWhiteSpace(Settings.ProxyUrl))
        {
            Settings.ProxyUrl = "http://127.0.0.1:10808";
        }
    }

    public void Save()
    {
        try
        {
            var json = JsonSerializer.Serialize(Settings, JsonOptions);
            File.WriteAllText(_settingsFilePath, json);
        }
        catch
        {
            // Ignore error
        }
    }
}
