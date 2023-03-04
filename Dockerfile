FROM mcr.microsoft.com/dotnet/sdk:6.0 AS build-env
WORKDIR /app

# Copy and restore as distinct layers
COPY *.sln ./
COPY ./src/Ramiel.Bot/*.csproj ./src/Ramiel.Bot/
COPY ./src/Ramiel.Discord/*.csproj ./src/Ramiel.Discord/

RUN dotnet restore

# Copy everything else and build
COPY . ./
RUN find -type d -name bin -prune -exec rm -rf {} \; && find -type d -name obj -prune -exec rm -rf {} \;
RUN dotnet publish -c Release -o /app/out

# Build runtime image
FROM mcr.microsoft.com/dotnet/aspnet:6.0

# Copy the app
WORKDIR /app
COPY --from=build-env /app/out .

# Start the app
ENTRYPOINT dotnet Ramiel.Bot.dll