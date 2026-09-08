using System.Text.RegularExpressions;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;

namespace IPAToolGUI.Views;

public partial class TwoFactorAuthDialog : Window
{
    public string AuthCode { get; private set; } = string.Empty;

    public TwoFactorAuthDialog()
    {
        InitializeComponent();
        Loaded += (_, _) =>
        {
            CodeTextBox.Focus();
        };
    }

    private void CodeTextBox_TextChanged(object sender, TextChangedEventArgs e)
    {
        var text = CodeTextBox.Text.Trim();
        ConfirmButton.IsEnabled = text.Length == 6 && Regex.IsMatch(text, @"^\d{6}$");
    }

    private void CodeTextBox_KeyDown(object sender, KeyEventArgs e)
    {
        if (e.Key == Key.Enter && ConfirmButton.IsEnabled)
        {
            Confirm();
        }
        else if (e.Key == Key.Escape)
        {
            DialogResult = false;
            Close();
        }
    }

    private void ConfirmButton_Click(object sender, RoutedEventArgs e)
    {
        Confirm();
    }

    private void Confirm()
    {
        var text = CodeTextBox.Text.Trim();
        if (text.Length != 6 || !Regex.IsMatch(text, @"^\d{6}$"))
        {
            MessageBox.Show("请输入正确的 6 位数字验证码！", "提示", MessageBoxButton.OK, MessageBoxImage.Warning);
            return;
        }

        AuthCode = text;
        DialogResult = true;
        Close();
    }

    private void CancelButton_Click(object sender, RoutedEventArgs e)
    {
        DialogResult = false;
        Close();
    }
}
