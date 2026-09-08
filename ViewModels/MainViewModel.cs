using System.Collections.ObjectModel;
using System.Diagnostics;
using System.IO;
using System.Windows;
using System.Windows.Data;
using System.Windows.Threading;
using IPAToolGUI.Models;
using IPAToolGUI.Services;

namespace IPAToolGUI.ViewModels;

public class MainViewModel : ViewModelBase
{
    private readonly SettingsService _settingsService;
    private readonly IpaToolService _ipaToolService;
    private readonly Dispatcher _dispatcher;
    private CancellationTokenSource? _cts;

    public MainViewModel()
    {
        _dispatcher = Application.Current?.Dispatcher ?? Dispatcher.CurrentDispatcher;
        _settingsService = new SettingsService();
        _ipaToolService = new IpaToolService(_settingsService);

        _ipaToolService.LogReceived += OnLogReceived;

        // Init settings binding
        SettingKeychainPassphrase = _settingsService.Settings.KeychainPassphrase;
        SettingDownloadDir = _settingsService.Settings.DefaultDownloadDir;
        SettingPlatform = _settingsService.Settings.DefaultPlatform;
        SettingEnableProxy = _settingsService.Settings.EnableProxy;
        SettingProxyUrl = _settingsService.Settings.ProxyUrl;

        DownloadOutputDir = _settingsService.Settings.DefaultDownloadDir;
        DownloadPlatform = _settingsService.Settings.DefaultPlatform;
        SearchPlatform = _settingsService.Settings.DefaultPlatform;

        // Init Collection Views
        _filteredVersionItemsView = CollectionViewSource.GetDefaultView(VersionItems);
        _filteredVersionItemsView.Filter = FilterVersions;

        // Commands
        CancelCommand = new RelayCommand(CancelOperation, () => IsBusy);
        ClearLogsCommand = new RelayCommand(() => Logs = string.Empty);

        RefreshAccountCommand = new AsyncRelayCommand(RefreshAccountAsync, () => !IsBusy);
        LoginCommand = new AsyncRelayCommand(LoginAsync, () => !IsBusy);
        RevokeCommand = new AsyncRelayCommand(RevokeAsync, () => !IsBusy && IsLoggedIn);
        ClearKeychainCacheCommand = new RelayCommand(ClearKeychainCache, () => !IsBusy);

        SearchCommand = new AsyncRelayCommand(SearchAsync, () => !IsBusy && !string.IsNullOrWhiteSpace(SearchTerm));
        ViewVersionsFromSearchCommand = new RelayCommand(p => ViewVersionsFromApp(p as AppItem));
        DownloadFromSearchCommand = new RelayCommand(p => DownloadFromApp(p as AppItem));
        PurchaseFromSearchCommand = new AsyncRelayCommand(PurchaseSelectedAppAsync, () => !IsBusy && SelectedSearchApp != null);

        ListVersionsCommand = new AsyncRelayCommand(ListVersionsAsync, () => !IsBusy && (!string.IsNullOrWhiteSpace(VersionTargetBundleId) || VersionTargetAppId > 0));
        QuerySelectedVersionMetadataCommand = new AsyncRelayCommand(QuerySelectedVersionMetadataAsync, () => !IsBusy && SelectedVersionItem != null);
        DownloadSelectedVersionCommand = new RelayCommand(DownloadSelectedVersion);
        BatchQueryMetadataCommand = new AsyncRelayCommand(BatchQueryMetadataAsync, () => !IsBusy && VersionItems.Count > 0);

        StartDownloadCommand = new AsyncRelayCommand(StartDownloadAsync, () => !IsBusy && (!string.IsNullOrWhiteSpace(DownloadBundleId) || DownloadAppId > 0));
        OpenDownloadDirCommand = new RelayCommand(OpenDownloadDir);
        SelectOutputDirCommand = new RelayCommand(SelectOutputDir);

        RefreshPurchasesCommand = new AsyncRelayCommand(RefreshPurchasesAsync, () => !IsBusy && IsLoggedIn);
        NextPurchasesPageCommand = new AsyncRelayCommand(NextPurchasesPageAsync, () => !IsBusy && (PurchasedPage * PurchasedLimit < PurchasedTotalCount));
        PrevPurchasesPageCommand = new AsyncRelayCommand(PrevPurchasesPageAsync, () => !IsBusy && PurchasedPage > 1);

        SaveSettingsCommand = new RelayCommand(SaveSettings);
        TestProxyCommand = new AsyncRelayCommand(TestProxyAsync);
        BrowseSettingDownloadDirCommand = new RelayCommand(BrowseSettingDownloadDir);

        // Auto check account status on startup
        _ = Task.Run(async () =>
        {
            await Task.Delay(300);
            await RefreshAccountAsync();
        });
    }

    #region Common State

    private int _selectedTabIndex;
    public int SelectedTabIndex
    {
        get => _selectedTabIndex;
        set => SetProperty(ref _selectedTabIndex, value);
    }

    private bool _isBusy;
    public bool IsBusy
    {
        get => _isBusy;
        set
        {
            if (SetProperty(ref _isBusy, value))
            {
                OnPropertyChanged(nameof(CanOperate));
            }
        }
    }

    public bool CanOperate => !IsBusy;

    private string _statusMessage = "就绪";
    public string StatusMessage
    {
        get => _statusMessage;
        set => SetProperty(ref _statusMessage, value);
    }

    private string _logs = string.Empty;
    public string Logs
    {
        get => _logs;
        set => SetProperty(ref _logs, value);
    }

    private void OnLogReceived(string log)
    {
        _dispatcher.Invoke(() =>
        {
            Logs += log + Environment.NewLine;
            // Cap logs at 50,000 characters
            if (Logs.Length > 50000)
            {
                Logs = Logs.Substring(Logs.Length - 40000);
            }
        });
    }

    public RelayCommand CancelCommand { get; }
    public RelayCommand ClearLogsCommand { get; }

    private void CancelOperation()
    {
        _cts?.Cancel();
    }

    #endregion

    #region Tab 1: Account

    private string _accountName = "未登录";
    public string AccountName
    {
        get => _accountName;
        set => SetProperty(ref _accountName, value);
    }

    private string _accountEmail = string.Empty;
    public string AccountEmail
    {
        get => _accountEmail;
        set => SetProperty(ref _accountEmail, value);
    }

    private bool _isLoggedIn;
    public bool IsLoggedIn
    {
        get => _isLoggedIn;
        set => SetProperty(ref _isLoggedIn, value);
    }

    private string _loginEmail = string.Empty;
    public string LoginEmail
    {
        get => _loginEmail;
        set => SetProperty(ref _loginEmail, value);
    }

    private string _loginPassword = string.Empty;
    public string LoginPassword
    {
        get => _loginPassword;
        set => SetProperty(ref _loginPassword, value);
    }

    private string _login2FA = string.Empty;
    public string Login2FA
    {
        get => _login2FA;
        set => SetProperty(ref _login2FA, value);
    }

    public AsyncRelayCommand RefreshAccountCommand { get; }
    public AsyncRelayCommand LoginCommand { get; }
    public AsyncRelayCommand RevokeCommand { get; }
    public RelayCommand ClearKeychainCacheCommand { get; }

    private async Task RefreshAccountAsync()
    {
        IsBusy = true;
        StatusMessage = "正在获取账号信息...";
        _cts = new CancellationTokenSource();

        try
        {
            var (success, account, error) = await _ipaToolService.GetAccountInfoAsync(_cts.Token);
            if (success && account != null)
            {
                AccountName = account.Name;
                AccountEmail = account.Email;
                IsLoggedIn = true;
                StatusMessage = $"已登录: {account.Name} ({account.Email})";
            }
            else
            {
                AccountName = "未登录";
                AccountEmail = string.Empty;
                IsLoggedIn = false;
                StatusMessage = string.IsNullOrWhiteSpace(error) ? "未检测到已登录的 Apple ID" : $"未登录: {error}";
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private async Task LoginAsync()
    {
        if (string.IsNullOrWhiteSpace(LoginEmail) || string.IsNullOrWhiteSpace(LoginPassword))
        {
            MessageBox.Show("请输入 Apple ID 邮箱和密码！", "提示", MessageBoxButton.OK, MessageBoxImage.Warning);
            return;
        }

        IsBusy = true;
        StatusMessage = "正在验证 Apple ID 账号与密码...";
        _cts = new CancellationTokenSource();

        try
        {
            var res = await _ipaToolService.LoginAsync(LoginEmail, LoginPassword, authCode: "", _cts.Token);

            // If 2FA verification is required by Apple, popup dialog for user input
            if (res.Requires2FA)
            {
                IsBusy = false;
                StatusMessage = "等待输入双重认证验证码...";

                var dialog = new Views.TwoFactorAuthDialog
                {
                    Owner = Application.Current?.MainWindow
                };

                if (dialog.ShowDialog() != true || string.IsNullOrWhiteSpace(dialog.AuthCode))
                {
                    StatusMessage = "用户取消了双重认证验证。";
                    return;
                }

                IsBusy = true;
                StatusMessage = "正在提交 2FA 验证码并完成登录...";
                _cts = new CancellationTokenSource();
                res = await _ipaToolService.LoginAsync(LoginEmail, LoginPassword, dialog.AuthCode, _cts.Token);
            }

            if (res.Success && res.Account != null)
            {
                AccountName = res.Account.Name;
                AccountEmail = res.Account.Email;
                IsLoggedIn = true;
                StatusMessage = $"登录成功: {res.Account.Name}";
                LoginPassword = string.Empty;
                MessageBox.Show($"登录成功！\n用户: {res.Account.Name}\n邮箱: {res.Account.Email}", "提示", MessageBoxButton.OK, MessageBoxImage.Information);
            }
            else
            {
                StatusMessage = $"登录失败: {res.ErrorMessage}";
                MessageBox.Show($"登录失败: {res.ErrorMessage}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private async Task RevokeAsync()
    {
        if (MessageBox.Show("确定要退出当前账号并清除登录凭据吗？", "确认", MessageBoxButton.YesNo, MessageBoxImage.Question) != MessageBoxResult.Yes)
        {
            return;
        }

        IsBusy = true;
        StatusMessage = "正在登出...";
        _cts = new CancellationTokenSource();

        try
        {
            var (success, error) = await _ipaToolService.RevokeAsync(_cts.Token);
            if (success)
            {
                AccountName = "未登录";
                AccountEmail = string.Empty;
                IsLoggedIn = false;
                StatusMessage = "已退出登录";
                MessageBox.Show("已成功注销登录凭据。", "提示", MessageBoxButton.OK, MessageBoxImage.Information);
            }
            else
            {
                StatusMessage = $"注销失败: {error}";
                MessageBox.Show($"注销失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private void ClearKeychainCache()
    {
        if (MessageBox.Show("此操作将删除本地用户目录下的 .ipatool 密钥文件夹 (%USERPROFILE%\\.ipatool)，重置所有本地缓存。\n确定要清理吗？", "确认清理", MessageBoxButton.YesNo, MessageBoxImage.Warning) != MessageBoxResult.Yes)
        {
            return;
        }

        try
        {
            var path = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".ipatool");
            if (Directory.Exists(path))
            {
                Directory.Delete(path, true);
                MessageBox.Show("本地密钥缓存已彻底清除！", "成功", MessageBoxButton.OK, MessageBoxImage.Information);
                _ = RefreshAccountAsync();
            }
            else
            {
                MessageBox.Show("未找到本地 .ipatool 缓存目录，无需清理。", "提示", MessageBoxButton.OK, MessageBoxImage.Information);
            }
        }
        catch (Exception ex)
        {
            MessageBox.Show($"清理失败: {ex.Message}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
        }
    }

    #endregion

    #region Tab 2: Search

    private string _searchTerm = "支付宝";
    public string SearchTerm
    {
        get => _searchTerm;
        set => SetProperty(ref _searchTerm, value);
    }

    private int _searchLimit = 10;
    public int SearchLimit
    {
        get => _searchLimit;
        set => SetProperty(ref _searchLimit, value);
    }

    private string _searchPlatform = "iphone";
    public string SearchPlatform
    {
        get => _searchPlatform;
        set => SetProperty(ref _searchPlatform, value);
    }

    public ObservableCollection<AppItem> SearchResults { get; } = new();

    private AppItem? _selectedSearchApp;
    public AppItem? SelectedSearchApp
    {
        get => _selectedSearchApp;
        set => SetProperty(ref _selectedSearchApp, value);
    }

    public AsyncRelayCommand SearchCommand { get; }
    public RelayCommand ViewVersionsFromSearchCommand { get; }
    public RelayCommand DownloadFromSearchCommand { get; }
    public AsyncRelayCommand PurchaseFromSearchCommand { get; }

    private async Task SearchAsync()
    {
        IsBusy = true;
        StatusMessage = $"正在搜索 \"{SearchTerm}\"...";
        _cts = new CancellationTokenSource();

        try
        {
            SearchResults.Clear();
            var (success, apps, error) = await _ipaToolService.SearchAsync(SearchTerm, SearchLimit, SearchPlatform, _cts.Token);
            if (success)
            {
                foreach (var app in apps)
                {
                    SearchResults.Add(app);
                }
                StatusMessage = $"搜索完成，找到 {apps.Count} 个应用。";
            }
            else
            {
                StatusMessage = $"搜索失败: {error}";
                MessageBox.Show($"搜索失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private void ViewVersionsFromApp(AppItem? app)
    {
        if (app == null) return;
        VersionTargetBundleId = app.BundleId;
        VersionTargetAppId = app.Id;
        SelectedTabIndex = 2; // Switch to Versions tab
        _ = ListVersionsAsync();
    }

    private void DownloadFromApp(AppItem? app)
    {
        if (app == null) return;
        DownloadBundleId = app.BundleId;
        DownloadAppId = app.Id;
        DownloadExternalVersionId = string.Empty; // Latest
        SelectedTabIndex = 3; // Switch to Download tab
    }

    private async Task PurchaseSelectedAppAsync()
    {
        if (SelectedSearchApp == null) return;
        IsBusy = true;
        StatusMessage = $"正在获取许可: {SelectedSearchApp.Name}...";
        _cts = new CancellationTokenSource();

        try
        {
            var (success, purchase, error) = await _ipaToolService.PurchaseAsync(SelectedSearchApp.BundleId, _cts.Token);
            if (success)
            {
                var msg = purchase?.AlreadyOwned == true ? "你已经拥有该应用的许可，无需重复购买。" : "获取免费许可成功！";
                StatusMessage = msg;
                MessageBox.Show(msg, "提示", MessageBoxButton.OK, MessageBoxImage.Information);
            }
            else
            {
                StatusMessage = $"获取许可失败: {error}";
                MessageBox.Show($"获取许可失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    #endregion

    #region Tab 3: History Versions

    private string _versionTargetBundleId = "com.alipay.iphoneclient";
    public string VersionTargetBundleId
    {
        get => _versionTargetBundleId;
        set => SetProperty(ref _versionTargetBundleId, value);
    }

    private long _versionTargetAppId;
    public long VersionTargetAppId
    {
        get => _versionTargetAppId;
        set => SetProperty(ref _versionTargetAppId, value);
    }

    private string _filterVersionText = string.Empty;
    public string FilterVersionText
    {
        get => _filterVersionText;
        set
        {
            if (SetProperty(ref _filterVersionText, value))
            {
                _filteredVersionItemsView.Refresh();
            }
        }
    }

    public ObservableCollection<AppVersionItem> VersionItems { get; } = new();
    private readonly System.ComponentModel.ICollectionView _filteredVersionItemsView;

    private bool FilterVersions(object obj)
    {
        if (string.IsNullOrWhiteSpace(FilterVersionText)) return true;
        if (obj is AppVersionItem item)
        {
            return item.VersionId.Contains(FilterVersionText, StringComparison.OrdinalIgnoreCase)
                || item.DisplayVersion.Contains(FilterVersionText, StringComparison.OrdinalIgnoreCase)
                || item.ReleaseDate.Contains(FilterVersionText, StringComparison.OrdinalIgnoreCase);
        }
        return true;
    }

    private AppVersionItem? _selectedVersionItem;
    public AppVersionItem? SelectedVersionItem
    {
        get => _selectedVersionItem;
        set => SetProperty(ref _selectedVersionItem, value);
    }

    public AsyncRelayCommand ListVersionsCommand { get; }
    public AsyncRelayCommand QuerySelectedVersionMetadataCommand { get; }
    public RelayCommand DownloadSelectedVersionCommand { get; }
    public AsyncRelayCommand BatchQueryMetadataCommand { get; }

    private async Task ListVersionsAsync()
    {
        IsBusy = true;
        StatusMessage = $"正在查询 {VersionTargetBundleId} 的历史版本构建 ID 列表...";
        _cts = new CancellationTokenSource();

        try
        {
            VersionItems.Clear();
            var (success, versionIds, error) = await _ipaToolService.ListVersionsAsync(VersionTargetBundleId, VersionTargetAppId, _cts.Token);
            if (success)
            {
                // App Store lists versions in ascending order, reverse to show newest first
                versionIds.Reverse();
                foreach (var id in versionIds)
                {
                    VersionItems.Add(new AppVersionItem { VersionId = id });
                }
                StatusMessage = $"共获取到 {versionIds.Count} 个历史版本构建 ID。";
            }
            else
            {
                StatusMessage = $"查询历史版本失败: {error}";
                MessageBox.Show($"查询历史版本失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private async Task QuerySelectedVersionMetadataAsync()
    {
        if (SelectedVersionItem == null) return;
        await ResolveVersionMetadataAsync(SelectedVersionItem);
    }

    private async Task ResolveVersionMetadataAsync(AppVersionItem item)
    {
        IsBusy = true;
        StatusMessage = $"正在查询版本 ID {item.VersionId} 的版本详情...";
        _cts = new CancellationTokenSource();

        try
        {
            item.IsQuerying = true;
            var (success, meta, error) = await _ipaToolService.GetVersionMetadataAsync(VersionTargetBundleId, item.VersionId, VersionTargetAppId, _cts.Token);
            if (success && meta != null)
            {
                item.DisplayVersion = meta.DisplayVersion;
                item.ReleaseDate = meta.ReleaseDate;
                _filteredVersionItemsView.Refresh();
                StatusMessage = $"版本 ID {item.VersionId} 对应版本号: {meta.DisplayVersion} (发布日期: {meta.ReleaseDate})";
            }
            else
            {
                StatusMessage = $"查询版本详情失败: {error}";
            }
        }
        finally
        {
            item.IsQuerying = false;
            IsBusy = false;
        }
    }

    private async Task BatchQueryMetadataAsync()
    {
        if (VersionItems.Count == 0) return;
        var countToQuery = Math.Min(VersionItems.Count, 30);
        if (MessageBox.Show($"即将批量查询前 {countToQuery} 个版本的详细版本号（每秒约查询 2 个），是否继续？", "批量查询确认", MessageBoxButton.YesNo, MessageBoxImage.Question) != MessageBoxResult.Yes)
        {
            return;
        }

        IsBusy = true;
        _cts = new CancellationTokenSource();

        try
        {
            for (int i = 0; i < countToQuery; i++)
            {
                if (_cts.IsCancellationRequested) break;
                var item = VersionItems[i];
                if (item.DisplayVersion != "未查询") continue;

                StatusMessage = $"批量查询中 ({i + 1}/{countToQuery}): {item.VersionId}...";
                var (success, meta, _) = await _ipaToolService.GetVersionMetadataAsync(VersionTargetBundleId, item.VersionId, VersionTargetAppId, _cts.Token);
                if (success && meta != null)
                {
                    item.DisplayVersion = meta.DisplayVersion;
                    item.ReleaseDate = meta.ReleaseDate;
                    _filteredVersionItemsView.Refresh();
                }

                await Task.Delay(200, _cts.Token);
            }
            StatusMessage = "批量版本查询完成。";
        }
        catch (OperationCanceledException)
        {
            StatusMessage = "批量查询已中止。";
        }
        finally
        {
            IsBusy = false;
        }
    }

    private void DownloadSelectedVersion(object? obj)
    {
        var item = obj as AppVersionItem ?? SelectedVersionItem;
        if (item == null) return;

        DownloadBundleId = VersionTargetBundleId;
        DownloadAppId = VersionTargetAppId;
        DownloadExternalVersionId = item.VersionId;
        SelectedTabIndex = 3; // Switch to Download tab
    }

    #endregion

    #region Tab 4: Download Center

    private string _downloadBundleId = "com.alipay.iphoneclient";
    public string DownloadBundleId
    {
        get => _downloadBundleId;
        set => SetProperty(ref _downloadBundleId, value);
    }

    private long _downloadAppId;
    public long DownloadAppId
    {
        get => _downloadAppId;
        set => SetProperty(ref _downloadAppId, value);
    }

    private string _downloadExternalVersionId = string.Empty;
    public string DownloadExternalVersionId
    {
        get => _downloadExternalVersionId;
        set => SetProperty(ref _downloadExternalVersionId, value);
    }

    private string _downloadOutputDir = string.Empty;
    public string DownloadOutputDir
    {
        get => _downloadOutputDir;
        set => SetProperty(ref _downloadOutputDir, value);
    }

    private string _downloadPlatform = "iphone";
    public string DownloadPlatform
    {
        get => _downloadPlatform;
        set => SetProperty(ref _downloadPlatform, value);
    }

    private bool _autoPurchaseLicense = true;
    public bool AutoPurchaseLicense
    {
        get => _autoPurchaseLicense;
        set => SetProperty(ref _autoPurchaseLicense, value);
    }

    private string _lastDownloadedFile = string.Empty;
    public string LastDownloadedFile
    {
        get => _lastDownloadedFile;
        set => SetProperty(ref _lastDownloadedFile, value);
    }

    public AsyncRelayCommand StartDownloadCommand { get; }
    public RelayCommand OpenDownloadDirCommand { get; }
    public RelayCommand SelectOutputDirCommand { get; }

    private async Task StartDownloadAsync()
    {
        IsBusy = true;
        var verLabel = string.IsNullOrWhiteSpace(DownloadExternalVersionId) ? "最新版本" : $"指定版本 ID: {DownloadExternalVersionId}";
        StatusMessage = $"正在下载 {DownloadBundleId} ({verLabel})...";
        _cts = new CancellationTokenSource();

        try
        {
            var outputPath = DownloadOutputDir;
            var (success, download, error) = await _ipaToolService.DownloadAsync(
                DownloadBundleId,
                DownloadAppId,
                DownloadExternalVersionId,
                outputPath,
                DownloadPlatform,
                AutoPurchaseLicense,
                _cts.Token);

            if (success && download != null)
            {
                LastDownloadedFile = download.Output;
                StatusMessage = $"下载成功！文件路径: {download.Output}";
                var result = MessageBox.Show($"下载成功！\n保存路径: {download.Output}\n\n是否立即打开所在文件夹？", "下载完成", MessageBoxButton.YesNo, MessageBoxImage.Information);
                if (result == MessageBoxResult.Yes)
                {
                    OpenFolderAndSelectFile(download.Output);
                }
            }
            else
            {
                StatusMessage = $"下载失败: {error}";
                MessageBox.Show($"下载失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private void SelectOutputDir()
    {
        var dialog = new Microsoft.Win32.OpenFolderDialog
        {
            Title = "选择 IPA 文件保存目录",
            InitialDirectory = Directory.Exists(DownloadOutputDir) ? DownloadOutputDir : Environment.GetFolderPath(Environment.SpecialFolder.UserProfile)
        };

        if (dialog.ShowDialog() == true)
        {
            DownloadOutputDir = dialog.FolderName;
        }
    }

    private void OpenDownloadDir()
    {
        var dir = DownloadOutputDir;
        if (File.Exists(LastDownloadedFile))
        {
            OpenFolderAndSelectFile(LastDownloadedFile);
            return;
        }

        if (!Directory.Exists(dir))
        {
            dir = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
        }
        Process.Start(new ProcessStartInfo("explorer.exe", dir) { UseShellExecute = true });
    }

    private static void OpenFolderAndSelectFile(string filePath)
    {
        if (File.Exists(filePath))
        {
            Process.Start(new ProcessStartInfo("explorer.exe", $"/select,\"{filePath}\"") { UseShellExecute = true });
        }
        else
        {
            var dir = Path.GetDirectoryName(filePath);
            if (Directory.Exists(dir))
            {
                Process.Start(new ProcessStartInfo("explorer.exe", dir) { UseShellExecute = true });
            }
        }
    }

    #endregion

    #region Tab 5: Purchased Apps

    public ObservableCollection<AppItem> PurchasedApps { get; } = new();

    private int _purchasedPage = 1;
    public int PurchasedPage
    {
        get => _purchasedPage;
        set => SetProperty(ref _purchasedPage, value);
    }

    private int _purchasedLimit = 20;
    public int PurchasedLimit
    {
        get => _purchasedLimit;
        set => SetProperty(ref _purchasedLimit, value);
    }

    private int _purchasedTotalCount;
    public int PurchasedTotalCount
    {
        get => _purchasedTotalCount;
        set => SetProperty(ref _purchasedTotalCount, value);
    }

    public AsyncRelayCommand RefreshPurchasesCommand { get; }
    public AsyncRelayCommand NextPurchasesPageCommand { get; }
    public AsyncRelayCommand PrevPurchasesPageCommand { get; }

    private async Task RefreshPurchasesAsync()
    {
        IsBusy = true;
        StatusMessage = $"正在加载第 {PurchasedPage} 页已购应用列表...";
        _cts = new CancellationTokenSource();

        try
        {
            PurchasedApps.Clear();
            var (success, apps, total, error) = await _ipaToolService.ListPurchasesAsync(PurchasedPage, PurchasedLimit, _cts.Token);
            if (success)
            {
                PurchasedTotalCount = total;
                foreach (var app in apps)
                {
                    PurchasedApps.Add(app);
                }
                StatusMessage = $"已购应用加载完成 (第 {PurchasedPage} 页，共 {total} 个)。";
            }
            else
            {
                StatusMessage = $"加载已购列表失败: {error}";
                MessageBox.Show($"加载已购列表失败: {error}", "错误", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }
        finally
        {
            IsBusy = false;
        }
    }

    private async Task NextPurchasesPageAsync()
    {
        PurchasedPage++;
        await RefreshPurchasesAsync();
    }

    private async Task PrevPurchasesPageAsync()
    {
        if (PurchasedPage > 1)
        {
            PurchasedPage--;
            await RefreshPurchasesAsync();
        }
    }

    #endregion

    #region Tab 6: Settings

    public string ResolvedIpaToolPath => SettingsService.ResolveIpaToolPath();

    private string _settingKeychainPassphrase = string.Empty;
    public string SettingKeychainPassphrase
    {
        get => _settingKeychainPassphrase;
        set => SetProperty(ref _settingKeychainPassphrase, value);
    }

    private string _settingDownloadDir = string.Empty;
    public string SettingDownloadDir
    {
        get => _settingDownloadDir;
        set => SetProperty(ref _settingDownloadDir, value);
    }

    private string _settingPlatform = "iphone";
    public string SettingPlatform
    {
        get => _settingPlatform;
        set => SetProperty(ref _settingPlatform, value);
    }

    private bool _settingEnableProxy = true;
    public bool SettingEnableProxy
    {
        get => _settingEnableProxy;
        set => SetProperty(ref _settingEnableProxy, value);
    }

    private string _settingProxyUrl = "http://127.0.0.1:10808";
    public string SettingProxyUrl
    {
        get => _settingProxyUrl;
        set => SetProperty(ref _settingProxyUrl, value);
    }

    public RelayCommand SaveSettingsCommand { get; }
    public AsyncRelayCommand TestProxyCommand { get; }
    public RelayCommand BrowseSettingDownloadDirCommand { get; }

    private void BrowseSettingDownloadDir()
    {
        var dialog = new Microsoft.Win32.OpenFolderDialog
        {
            Title = "选择默认下载保存目录",
            InitialDirectory = Directory.Exists(SettingDownloadDir) ? SettingDownloadDir : Environment.GetFolderPath(Environment.SpecialFolder.UserProfile)
        };

        if (dialog.ShowDialog() == true)
        {
            SettingDownloadDir = dialog.FolderName;
        }
    }

    private async Task TestProxyAsync()
    {
        if (string.IsNullOrWhiteSpace(SettingProxyUrl))
        {
            MessageBox.Show("请输入代理地址！", "提示", MessageBoxButton.OK, MessageBoxImage.Warning);
            return;
        }

        IsBusy = true;
        StatusMessage = "正在测试网络代理连接...";
        try
        {
            var handler = new System.Net.Http.HttpClientHandler
            {
                Proxy = new System.Net.WebProxy(SettingProxyUrl),
                UseProxy = true
            };
            using var client = new System.Net.Http.HttpClient(handler) { Timeout = TimeSpan.FromSeconds(6) };
            var res = await client.GetAsync("https://apple.com");
            StatusMessage = "网络代理测试成功！";
            MessageBox.Show($"代理连接测试成功！\n目标: https://apple.com\n响应码: {(int)res.StatusCode} {res.StatusCode}", "连接成功", MessageBoxButton.OK, MessageBoxImage.Information);
        }
        catch (Exception ex)
        {
            StatusMessage = $"代理连接测试失败: {ex.Message}";
            MessageBox.Show($"代理连接测试失败: {ex.Message}\n请检查代理软件是否开启本地监听端口（如 10808）。", "测试失败", MessageBoxButton.OK, MessageBoxImage.Error);
        }
        finally
        {
            IsBusy = false;
        }
    }

    private void SaveSettings()
    {
        _settingsService.Settings.KeychainPassphrase = SettingKeychainPassphrase;
        _settingsService.Settings.DefaultDownloadDir = SettingDownloadDir;
        _settingsService.Settings.DefaultPlatform = SettingPlatform;
        _settingsService.Settings.EnableProxy = SettingEnableProxy;
        _settingsService.Settings.ProxyUrl = SettingProxyUrl;
        _settingsService.Save();

        DownloadOutputDir = SettingDownloadDir;
        DownloadPlatform = SettingPlatform;

        MessageBox.Show("设置已保存并生效！", "提示", MessageBoxButton.OK, MessageBoxImage.Information);
    }

    #endregion
}
