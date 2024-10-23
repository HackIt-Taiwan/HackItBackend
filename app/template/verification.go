package htmlTempalte

var VerificationTemplate string = `
<!DOCTYPE html>
<html lang="zh-Hant">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>請驗證您的信箱</title>
    <style>
      body {
        font-family: Arial, sans-serif;
        background-color: #f4f4f4;
        color: #333;
        margin: 0;
        padding: 0;
        display: flex;
        justify-content: center;
        align-items: center;
        height: 100vh;
        text-align: center;
      }

      .container {
        background: white;
        padding: 20px;
        border-radius: 8px;
        box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
        max-width: 400px;
        margin: auto;
      }

      h1 {
        color: #007bff;
      }

      p {
        font-size: 16px;
        line-height: 1.5;
        margin: 15px 0;
      }

    </style>
  </head>
  <body>
    <div class="container">
      <h1>請驗證您的信箱</h1>
	<p style="color: black;">
	親愛的{{.Name}}，感謝您的報名
	Hackit！為了確保您的帳戶安全，請驗證您的電子郵件地址。
	</p>
      <p>只需點擊下面的按鈕，即可完成驗證：</p>
<a href="{{.VerificationLink}}" 
   style="display: inline-block; padding: 10px 20px; font-size: 16px; color: white; background-color: #007bff; border: none; border-radius: 5px; text-decoration: none; transition: background-color 0.3s;">
   驗證我的信箱
</a>    </div>
  </body>
</html>
`
