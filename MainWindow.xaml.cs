using System.Windows;
using System.Windows.Controls;

namespace IPAToolGUI;

/// <summary>
/// Interaction logic for MainWindow.xaml
/// </summary>
public partial class MainWindow : Window
{
    public MainWindow()
    {
        InitializeComponent();
    }

    private void LogsTextBox_TextChanged(object sender, TextChangedEventArgs e)
    {
        if (sender is TextBox textBox)
        {
            textBox.CaretIndex = textBox.Text.Length;
            textBox.ScrollToEnd();
        }
    }
}
