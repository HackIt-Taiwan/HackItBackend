package htmlTempalte

var TeamLeaderEditTemplate string = `
<!DOCTYPE html>
<html lang="zh-Hant">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>對表單進行編輯</title>
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
      <h1>對表單進行編輯</h1>
	<p style="color: black;">
	親愛的{{.Name}}，感謝您的報名
	Hackit！這封郵件是傳送給你編輯的權限(因為你是團隊代表人)!
	</p>
      <p>只需點擊下面的按鈕，即可進入編輯頁面：</p>
<a href="{{.EditLink}}" 
   style="display: inline-block; padding: 10px 20px; font-size: 16px; color: white; background-color: #007bff; border: none; border-radius: 5px; text-decoration: none; transition: background-color 0.3s;">
   進入編輯頁面
</a>    </div>
  </body>
</html>
`
