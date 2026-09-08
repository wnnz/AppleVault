using System.Text.Json.Serialization;

namespace IPAToolGUI.Models;

public class AccountInfo
{
    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("email")]
    public string Email { get; set; } = string.Empty;

    [JsonPropertyName("success")]
    public bool Success { get; set; }
}

public class AppItem
{
    [JsonPropertyName("id")]
    public long Id { get; set; }

    [JsonPropertyName("bundleID")]
    public string BundleId { get; set; } = string.Empty;

    [JsonPropertyName("name")]
    public string Name { get; set; } = string.Empty;

    [JsonPropertyName("version")]
    public string Version { get; set; } = string.Empty;

    [JsonPropertyName("price")]
    public double Price { get; set; }

    [JsonPropertyName("purchaseDate")]
    public string? PurchaseDate { get; set; }

    public string DisplayPrice => Price == 0 ? "免费" : $"¥{Price:F2}";
}

public class SearchResult
{
    [JsonPropertyName("count")]
    public int Count { get; set; }

    [JsonPropertyName("apps")]
    public List<AppItem> Apps { get; set; } = new();
}

public class PurchasedResult
{
    [JsonPropertyName("count")]
    public int Count { get; set; }

    [JsonPropertyName("totalCount")]
    public int TotalCount { get; set; }

    [JsonPropertyName("page")]
    public int Page { get; set; }

    [JsonPropertyName("apps")]
    public List<AppItem> Apps { get; set; } = new();
}

public class VersionsResult
{
    [JsonPropertyName("bundleID")]
    public string BundleId { get; set; } = string.Empty;

    [JsonPropertyName("externalVersionIdentifiers")]
    public List<string> ExternalVersionIdentifiers { get; set; } = new();

    [JsonPropertyName("success")]
    public bool Success { get; set; }
}

public class VersionMetadataResult
{
    [JsonPropertyName("externalVersionID")]
    public string ExternalVersionId { get; set; } = string.Empty;

    [JsonPropertyName("displayVersion")]
    public string DisplayVersion { get; set; } = string.Empty;

    [JsonPropertyName("releaseDate")]
    public string ReleaseDate { get; set; } = string.Empty;

    [JsonPropertyName("success")]
    public bool Success { get; set; }
}

public class DownloadResult
{
    [JsonPropertyName("output")]
    public string Output { get; set; } = string.Empty;

    [JsonPropertyName("purchased")]
    public bool Purchased { get; set; }

    [JsonPropertyName("success")]
    public bool Success { get; set; }
}

public class PurchaseResult
{
    [JsonPropertyName("alreadyOwned")]
    public bool AlreadyOwned { get; set; }

    [JsonPropertyName("success")]
    public bool Success { get; set; }
}

public class AppVersionItem
{
    public string VersionId { get; set; } = string.Empty;
    public string DisplayVersion { get; set; } = "未查询";
    public string ReleaseDate { get; set; } = "-";
    public bool IsQuerying { get; set; }
}

public class AppSettings
{
    public string KeychainPassphrase { get; set; } = string.Empty;
    public string DefaultDownloadDir { get; set; } = string.Empty;
    public string DefaultPlatform { get; set; } = "iphone";
    public bool EnableProxy { get; set; } = true;
    public string ProxyUrl { get; set; } = "http://127.0.0.1:10808";
}
